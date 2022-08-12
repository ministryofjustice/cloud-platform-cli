package environment

import (
	"fmt"
	"log"

	"github.com/kelseyhightower/envconfig"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/util"
)

// Options are used to configure apply sessions.
// These options are normally passed via flags in a command line.
type Options struct {
	Namespace, KubecfgPath, ClusterCtx, GitToken string
	PRNumber                                     int
	AllNamespaces                                bool
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

func (a *Apply) Apply() error {

	re := RepoEnvironment{}
	err := re.mustBeInCloudPlatformEnvironments()
	if err != nil {
		return err
	}
	err = a.applyNamespace()
	if err != nil {
		return err
	}
	return nil
}

func (a *Apply) Plan() error {

	re := RepoEnvironment{}
	err := re.mustBeInCloudPlatformEnvironments()
	if err != nil {
		return err
	}

	changedNamespaces, err := util.ChangedInPR(a.Options.GitToken, cloudPlatformEnvRepo, mojOwner, a.Options.PRNumber)
	if err != nil {
		return err
	}

	for _, namespace := range changedNamespaces {
		a.Options.Namespace = namespace
		err = a.applyNamespace()
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Apply) ApplyAll() error {

	repoPath := "namespaces/" + a.Options.ClusterCtx
	folderChunks := util.GetFolderChunks(repoPath, numRoutines)

	for i := 0; i < len(folderChunks); i++ {
		err := a.applyNamespaceDirs(folderChunks[i])
		if err != nil {
			return err
		}

	}
	return nil
}

func (a *Apply) applyNamespaceDirs(chunkFolder []string) error {

	for _, folder := range chunkFolder {
		a.Options.Namespace = folder

		err := util.GetLatestGitPull()
		if err != nil {
			return err
		}
		err = a.applyNamespace()
		if err != nil {
			return err
		}

	}

	return nil
}

func (a *Apply) planKubectl() (string, error) {
	log.Printf("Doing kubectl dry-run for namespace: %v in directory %v", a.Options.Namespace, a.Dir)

	outputKubectl, err := a.Applier.KubectlApply(a.Options.Namespace, a.Dir, true)
	if err != nil {
		err := fmt.Errorf("error running kubectl on namespace %s: %v", a.Options.Namespace, err)
		return "", err
	}

	return outputKubectl, nil
}

func (a *Apply) applyKubectl() (string, error) {
	log.Printf("Apply kubectl for namespace: %v in directory %v", a.Options.Namespace, a.Dir)

	outputKubectl, err := a.Applier.KubectlApply(a.Options.Namespace, a.Dir, false)
	if err != nil {
		err := fmt.Errorf("error running kubectl on namespace %s: %v", a.Options.Namespace, err)
		return "", err
	}

	return outputKubectl, nil
}

func (a *Apply) planTerraform() (string, error) {
	log.Printf("Doing Terraform Plan for namespace: %v", a.Options.Namespace)

	tfFolder := a.Dir + "/resources"

	outputTerraform, err := a.Applier.TerraformInitAndPlan(a.Options.Namespace, tfFolder)
	if err != nil {
		err := fmt.Errorf("error running terraform on namespace %s: %v", a.Options.Namespace, err)
		return "", err
	}
	return outputTerraform, nil
}

func (a *Apply) applyTerraform() (string, error) {
	log.Printf("Applying Terraform for namespace: %v", a.Options.Namespace)

	tfFolder := a.Dir + "/resources"

	outputTerraform, err := a.Applier.TerraformInitAndApply(a.Options.Namespace, tfFolder)
	if err != nil {
		err := fmt.Errorf("error running terraform on namespace %s: %v", a.Options.Namespace, err)
		return "", err
	}
	return outputTerraform, nil
}

func (a *Apply) applyNamespace() error {

	applier, err := NewApply(*a.Options)
	if err != nil {
		return err
	}

	outputKubectl, err := applier.planKubectl()
	if err != nil {
		return err
	}

	outputTerraform, err := applier.planTerraform()
	if err != nil {
		return err
	}

	fmt.Println("\nOutput of kubectl:", outputKubectl, "\nOutput of terraform", outputTerraform)
	return nil

}
