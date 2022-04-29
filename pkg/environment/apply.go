package environment

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/kelseyhightower/envconfig"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/terraform"
	authenticate "github.com/ministryofjustice/cloud-platform-environments/pkg/authenticate"
)

type ConfigVars struct {
	PipelineStateBucket             string `required:"true" split_words:"true"`
	PipelineStateKeyPrefix          string `required:"true" split_words:"true"`
	PipelineTerraformStateLockTable string `required:"true" split_words:"true"`
	PipelineStateRegion             string `required:"true" split_words:"true"`
	PipelineCluster                 string `required:"true" split_words:"true"`
	PipelineClusterState            string `required:"true" split_words:"true"`
	Repo                            string `default:"cloud-platform-environments" split_words:"true"`
	Namespace                       string
}

type KubeConfig struct {
	KubeconfigS3Bucket string `required:"true" envconfig:"KUBECONFIG_S3_BUCKET"`
	KubeconfigS3Key    string `required:"true" envconfig:"KUBECONFIG_S3_KEY"`
	Context            string `default:"live.cloud-platform.service.justice.gov.uk"`
	AwsRegion          string `required:"true" split_words:"true"`
	Kubeconfig         string `default:"kubeconfig"`
}

type RequiredEnvConfig struct {
	clustername        string `required:"true" envconfig:"TF_VAR_cluster_name"`
	clusterstatebucket string `required:"true" envconfig:"TF_VAR_cluster_state_bucket"`
	clusterstatekey    string `required:"true" envconfig:"TF_VAR_cluster_state_key"`
	githubowner        string `required:"true" envconfig:"TF_VAR_github_owner"`
	githubtoken        string `required:"true" envconfig:"TF_VAR_github_token"`
	pingdomapitoken    string `required:"true" envconfig:"PINGDOM_API_TOKEN"`
}

func Apply(namespace string) error {

	var config ConfigVars
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatal(err.Error())
	}

	var kubecfg KubeConfig
	err = envconfig.Process("", &kubecfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	var tfConfig RequiredEnvConfig
	err = envconfig.Process("", &tfConfig)
	if err != nil {
		log.Fatal(err.Error())
	}

	re := RepoEnvironment{}
	err = re.mustBeInCloudPlatformEnvironments()
	if err != nil {
		return err
	}

	err = authenticate.SwitchContextFromS3Bucket(
		kubecfg.KubeconfigS3Bucket,
		kubecfg.KubeconfigS3Key,
		kubecfg.AwsRegion,
		kubecfg.Context,
		kubecfg.Kubeconfig)
	if err != nil {
		log.Fatalln("error in switching context", err)
	}

	err = ApplyNamespace(config)
	if err != nil {
		log.Fatal(err.Error())
	}
	return nil

}

func ApplyNamespace(config ConfigVars) error {
	log.Printf("Applying namespace: %v", config.Namespace)

	outputKubectl, err := applyKubectl(config)
	if err != nil {
		err := fmt.Errorf("error running kubectl on namespace %s: %v", config.Namespace, err)
		return err
	}

	key := config.PipelineStateKeyPrefix + config.PipelineClusterState + "/" + config.Namespace + "/terraform.tfstate"

	tfArgs := []string{
		"init",
		fmt.Sprintf("%s=bucket=%s", "-backend-config", config.PipelineStateBucket),
		fmt.Sprintf("%s=key=%s", "-backend-config", key),
		fmt.Sprintf("%s=dynamodb_table=%s", "-backend-config", config.PipelineTerraformStateLockTable),
		fmt.Sprintf("%s=region=%s", "-backend-config", config.PipelineStateRegion)}

	outputInitTf, err := runTerraform(config, tfArgs)
	if err != nil {
		err := fmt.Errorf("error running terraform init on namespace %s: %v: %v", config.Namespace, err.Error(), outputInitTf.Stderr)
		return err

	}

	tfArgs = []string{"apply"}
	outputPlanTf, err := runTerraform(config, tfArgs)
	if err != nil {
		err := fmt.Errorf("error running terraform plan  on namespace %s: %v: %v", config.Namespace, err.Error(), outputPlanTf.Stderr)
		return err

	}
	output := outputKubectl + "\n" + outputInitTf.Stdout + "\n" + outputPlanTf.Stdout

	fmt.Printf("Output of Namespace changes %s", output)
	return nil
}

// applyKubectl attempts to dryn-run of "kubectl apply" to the files in the given folder.
// It returns the apply command output and err.
func applyKubectl(config ConfigVars) (output string, err error) {

	kubectlArgs := []string{"-n", filepath.Base(config.Namespace), "apply", "-f", "."}

	//kubectlArgs = append(kubectlArgs, "--dry-run")

	kubectlCommand := exec.Command("kubectl", kubectlArgs...)

	kubectlCommand.Dir = "namespaces/" + config.PipelineCluster + "/" + config.Namespace
	log.Printf("RUN :  command %v on folder %v", kubectlCommand, config.Namespace)
	outb, err := kubectlCommand.Output()
	if err != nil {
		return "", err
	}

	return string(outb), nil

}

func runTerraform(config ConfigVars, tfArgs []string) (output *terraform.CmdOutput, err error) {

	Command := exec.Command("terraform", tfArgs...)

	Command.Dir = "namespaces/" + config.PipelineCluster + "/" + config.Namespace + "/resources"

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	var exitCode int

	Command.Stdout = &stdoutBuf
	Command.Stderr = &stderrBuf

	err = Command.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
		}
		cmdOutput := terraform.CmdOutput{
			Stdout:   stdoutBuf.String(),
			Stderr:   stderrBuf.String(),
			ExitCode: exitCode,
		}
		return &cmdOutput, err
	} else {
		ws := Command.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
	}

	cmdOutput := terraform.CmdOutput{
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
		ExitCode: exitCode,
	}

	if cmdOutput.ExitCode != 0 {
		return &cmdOutput, err
	} else {
		return &cmdOutput, nil
	}
}
