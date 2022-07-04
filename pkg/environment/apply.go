package environment

import (
	"log"
	"os"

	"github.com/kelseyhightower/envconfig"
	authenticate "github.com/ministryofjustice/cloud-platform-environments/pkg/authenticate"
)

type KubeConfig struct {
	KubeconfigS3Bucket string `required:"true" envconfig:"KUBECONFIG_S3_BUCKET"`
	KubeconfigS3Key    string `required:"true" envconfig:"KUBECONFIG_S3_KEY"`
	Context            string `default:"live.cloud-platform.service.justice.gov.uk"`
	AwsRegion          string `required:"true" split_words:"true"`
	Kubeconfig         string `default:"kubeconfig"`
}

type RequiredEnvVars struct {
	clustername        string `required:"true" envconfig:"TF_VAR_cluster_name"`
	clusterstatebucket string `required:"true" envconfig:"TF_VAR_cluster_state_bucket"`
	clusterstatekey    string `required:"true" envconfig:"TF_VAR_cluster_state_key"`
	githubowner        string `required:"true" envconfig:"TF_VAR_github_owner"`
	githubtoken        string `required:"true" envconfig:"TF_VAR_github_token"`
	pingdomapitoken    string `required:"true" envconfig:"PINGDOM_API_TOKEN"`
}

type Apply struct {
	KubeConfig      KubeConfig
	RequiredEnvVars RequiredEnvVars
	Terraformer     Terraformer
	Dir             string
	Namespace       string
}

func NewApplier(namespace string) (*Apply, error) {

	var kubecfg KubeConfig
	err := envconfig.Process("", &kubecfg)
	if err != nil {
		log.Fatalln("Kubeconfig environment variables not set:", err.Error())
	}

	var reqConfig RequiredEnvVars
	err = envconfig.Process("", &reqConfig)
	if err != nil {
		log.Fatalln("Pipeline environment variables not set:", err.Error())
	}

	workingDir, _ := os.Getwd()

	applier := Apply{
		KubeConfig:      kubecfg,
		RequiredEnvVars: reqConfig,
		Terraformer:     NewTerraformer("/usr/local/bin/terraform"),
		Namespace:       namespace,
		Dir:             workingDir + "/namespaces/" + kubecfg.Context + "/" + namespace,
	}
	applier.initialize()

	return &applier, nil
}

func (a *Apply) initialize() {

	err := authenticate.SwitchContextFromS3Bucket(
		a.KubeConfig.KubeconfigS3Bucket,
		a.KubeConfig.KubeconfigS3Key,
		a.KubeConfig.AwsRegion,
		a.KubeConfig.Context,
		a.KubeConfig.Kubeconfig)
	if err != nil {
		log.Fatalln("error in switching context", err)
	}

}

func (a *Apply) ApplyNamespace() (map[string]string, error) {
	log.Printf("Applying namespace: %v", a.Namespace)

	// outputKubectl, err := applyKubectl()
	// if err != nil {
	// 	err := fmt.Errorf("error running kubectl on namespace %s: %v", a.ConfigVars.Namespace, err)
	// 	return err
	// }

	outputTerraform, _ := a.Terraformer.TerraformInitAndApply(a.Namespace, a.Dir)
	return outputTerraform, nil
}

// applyKubectl attempts to dryn-run of "kubectl apply" to the files in the given folder.
// It returns the apply command output and err.
// func (a *Apply) applyKubectl() (output string, err error) {

// 	kubectlArgs := []string{"-n", filepath.Base(a.Namespace), "apply", "-f", "."}

// 	//kubectlArgs = append(kubectlArgs, "--dry-run")

// 	kubectlCommand := exec.Command("kubectl", kubectlArgs...)

// 	kubectlCommand.Dir = "namespaces/" + a.Terraformer.config.PipelineCluster + "/" + config.Namespace
// 	log.Printf("RUN :  command %v on folder %v", kubectlCommand, config.Namespace)
// 	outb, err := kubectlCommand.Output()
// 	if err != nil {
// 		return "", err
// 	}

// 	return string(outb), nil

// }

// func runTerraform(config ConfigVars, tfArgs []string) (output *terraform.CmdOutput, err error) {

// 	Command := exec.Command("terraform", tfArgs...)

// 	Command.Dir = "namespaces/" + config.PipelineCluster + "/" + config.Namespace + "/resources"

// 	var stdoutBuf bytes.Buffer
// 	var stderrBuf bytes.Buffer
// 	var exitCode int

// 	Command.Stdout = &stdoutBuf
// 	Command.Stderr = &stderrBuf

// 	err = Command.Run()
// 	if err != nil {
// 		if exitError, ok := err.(*exec.ExitError); ok {
// 			ws := exitError.Sys().(syscall.WaitStatus)
// 			exitCode = ws.ExitStatus()
// 		}
// 		cmdOutput := terraform.CmdOutput{
// 			Stdout:   stdoutBuf.String(),
// 			Stderr:   stderrBuf.String(),
// 			ExitCode: exitCode,
// 		}
// 		return &cmdOutput, err
// 	} else {
// 		ws := Command.ProcessState.Sys().(syscall.WaitStatus)
// 		exitCode = ws.ExitStatus()
// 	}

// 	cmdOutput := terraform.CmdOutput{
// 		Stdout:   stdoutBuf.String(),
// 		Stderr:   stderrBuf.String(),
// 		ExitCode: exitCode,
// 	}

// 	if cmdOutput.ExitCode != 0 {
// 		return &cmdOutput, err
// 	} else {
// 		return &cmdOutput, nil
// 	}
// }
