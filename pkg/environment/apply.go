package environment

import (
	"log"
	"os"

	"github.com/kelseyhightower/envconfig"
)

// Options are used to configure apply sessions.
// These options are normally passed via flags in a command line.
type Options struct {
	Namespace, KubecfgPath, ClusterCtx string
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
	Options         *Options
	RequiredEnvVars RequiredEnvVars
	Terraformer     Terraformer
	Dir             string
}

func NewApply(opt Options) (*Apply, error) {
	workingDir, _ := os.Getwd()
	var reqEnvVars RequiredEnvVars
	err := envconfig.Process("", &reqEnvVars)
	if err != nil {
		return nil, err
	}
	return &Apply{
		Options:         &opt,
		Terraformer:     NewTerraformer("/usr/local/bin/terraform"),
		Dir:             workingDir + "/namespaces/" + opt.ClusterCtx + "/" + opt.Namespace,
		RequiredEnvVars: reqEnvVars,
	}, nil
}

func (a *Apply) Apply() (map[string]string, error) {
	applier, err := NewApply(*a.Options)
	if err != nil {
		return nil, err
	}

	_, err = applier.ApplyNamespace()
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (a *Apply) ApplyNamespace() (map[string]string, error) {
	log.Printf("Applying namespace: %v", a.Options.Namespace)

	// outputKubectl, err := applyKubectl()
	// if err != nil {
	// 	err := fmt.Errorf("error running kubectl on namespace %s: %v", a.ConfigVars.Namespace, err)
	// 	return err
	// }

	outputTerraform, _ := a.Terraformer.TerraformInitAndApply(a.Options.Namespace, a.Dir)
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
