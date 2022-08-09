package environment

import (
	"fmt"
	"log"

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
	Applier         Applier
	Dir             string
}

func NewApply(opt Options) (*Apply, error) {
	var reqEnvVars RequiredEnvVars
	err := envconfig.Process("", &reqEnvVars)
	if err != nil {
		return nil, err
	}
	return &Apply{
		Options:         &opt,
		Applier:         NewApplier("/usr/local/bin/terraform", "/usr/local/bin/kubectl"),
		Dir:             "namespaces/" + opt.ClusterCtx + "/" + opt.Namespace,
		RequiredEnvVars: reqEnvVars,
	}, nil
}

func (a *Apply) Apply() (map[string]string, error) {
	applier, err := NewApply(*a.Options)
	if err != nil {
		return nil, err
	}

	outputKubectl, err := applier.ApplyKubectl()
	if err != nil {
		return nil, err
	}

	outputTerraform, err := applier.ApplyTerraform()
	if err != nil {
		return nil, err
	}

	fmt.Println("\nOutput of kubectl:", outputKubectl, "\nOutput of terraform", outputTerraform)
	return nil, nil
}

func (a *Apply) ApplyKubectl() (string, error) {
	log.Printf("Applying kubectl for namespace: %v in directory %v", a.Options.Namespace, a.Dir)

	outputKubectl, err := a.Applier.KubectlApply(a.Dir)
	if err != nil {
		err := fmt.Errorf("error running kubectl on namespace %s: %v", a.Options.Namespace, err)
		return "", err
	}

	return outputKubectl, nil
}

func (a *Apply) ApplyTerraform() (string, error) {
	log.Printf("Applying Terraform for namespace: %v", a.Options.Namespace)

	tfFolder := a.Dir + "/resources"

	outputTerraform, err := a.Applier.TerraformInitAndApply(a.Options.Namespace, tfFolder)
	if err != nil {
		err := fmt.Errorf("error running terraform on namespace %s: %v", a.Options.Namespace, err)
		return "", err
	}
	return outputTerraform, nil
}
