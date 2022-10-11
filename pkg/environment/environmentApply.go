package environment

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/util"
)

// Options are used to configure plan/apply sessions.
// These options are normally passed via flags in a command line.
type Options struct {
	Namespace, KubecfgPath, ClusterCtx, GithubToken string
	PRNumber                                        int
	AllNamespaces                                   bool
}

// RequiredEnvVars is used to store values such as TF_VAR_ , github and pingdom tokens
// which are needed to perform terraform operations for a given namespace
type RequiredEnvVars struct {
	clustername        string `required:"true" envconfig:"TF_VAR_cluster_name"`
	clusterstatebucket string `required:"true" envconfig:"TF_VAR_cluster_state_bucket"`
	kubernetescluster  string `required:"true" envconfig:"TF_VAR_kubernetes_cluster"`
	githubowner        string `required:"true" envconfig:"TF_VAR_github_owner"`
	githubtoken        string `required:"true" envconfig:"TF_VAR_github_token"`
	pingdomapitoken    string `required:"true" envconfig:"PINGDOM_API_TOKEN"`
}

// Apply is used to store objects in a Apply/Plan session
type Apply struct {
	Options         *Options
	RequiredEnvVars RequiredEnvVars
	Applier         Applier
	Dir             string
}

const (
	// Assumption that there are no more than 5 PRs merged in last minute
	prCount = 5
)

// NewApply creates a new Apply object and populates its fields with values from options(which are flags),
// instantiate Applier object which also checks and sets the Backend config variables to do terraform init,
// RequiredEnvVars object which stores the values required for plan/apply of namespace
func NewApply(opt Options) *Apply {
	apply := Apply{
		Options: &opt,
		Applier: NewApplier("/usr/local/bin/terraform", "/usr/local/bin/kubectl"),
		Dir:     "namespaces/" + opt.ClusterCtx + "/" + opt.Namespace,
	}

	apply.Initialize()
	return &apply
}

func (a *Apply) Initialize() {
	var reqEnvVars RequiredEnvVars
	err := envconfig.Process("", &reqEnvVars)
	if err != nil {
		log.Fatalln("Environment variables required to perform terraform operations not set:", err.Error())
	}
	a.RequiredEnvVars.clustername = reqEnvVars.clustername
	a.RequiredEnvVars.githubowner = reqEnvVars.githubowner
	a.RequiredEnvVars.githubtoken = reqEnvVars.githubtoken
	a.RequiredEnvVars.pingdomapitoken = reqEnvVars.pingdomapitoken
}

// Plan is the entry point for performing a namespace plan.
// It checks if the working directory is in cloud-platform-environments, checks if a PR number or a namespace is given
// If a namespace is given, it perform a kubectl apply -dry-run and a terraform init and plan of that namespace
// else checks for PR number and get the list of changed namespaces in the PR. Then does the kubectl apply -dry-run and
// terraform init and plan of all the namespaces changed in the PR
func (a *Apply) Plan() error {

	if a.Options.PRNumber == 0 && a.Options.Namespace == "" {
		err := fmt.Errorf("either a PR Id/Number or a namespace is required to perform plan")
		return err
	}

	// If a namespace is given as a flag, then perform a plan for the given namespace.
	if a.Options.Namespace != "" {
		err := a.planNamespace()
		if err != nil {
			return err
		}
		return nil
	} else {
		changedNamespaces, err := util.NSChangedInPR(a.Options.ClusterCtx, a.Options.GithubToken, cloudPlatformEnvRepo, mojOwner, a.Options.PRNumber)
		if err != nil {
			return err
		}

		if changedNamespaces != nil {
			for _, namespace := range changedNamespaces {
				a.Options.Namespace = namespace
				err = a.planNamespace()
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Apply is the entry point for performing a namespace apply.
// It checks if the working directory is in cloud-platform-environments, checks if a PR number or a namespace is given
// If a namespace is given, it perform a kubectl apply and a terraform init and apply of that namespace
// else checks for PR number and get the list of changed namespaces in that merged PR. Then does the kubectl apply and
// terraform init and apply of all the namespaces merged in the PR
func (a *Apply) Apply() error {

	if a.Options.PRNumber == 0 && a.Options.Namespace == "" {
		err := fmt.Errorf("either a PR Id/Number or a namespace is required to perform apply")
		return err
	}

	// If a namespace is given as a flag, then perform a apply for the given namespace.
	if a.Options.Namespace != "" {
		err := a.applyNamespace()
		if err != nil {
			return err
		}
	} else {
		// get the current and current - 1 minute
		date := util.GetDateLastMinute()
		// get the list of PRs that are merged in past 1 minute
		nodes, err := util.GetMergedPRs(date, prCount, a.Options.GithubToken)
		if err != nil {
			return err
		}

		for _, pr := range nodes {
			url := string(pr.PullRequest.Url)
			prNumber, err := strconv.Atoi(url[strings.LastIndex(url, "/")+1:])
			if err != nil {
				return err
			}
			changedNamespaces, err := util.NSChangedInPR(a.Options.ClusterCtx, a.Options.GithubToken, cloudPlatformEnvRepo, mojOwner, prNumber)
			if err != nil {
				return err
			}

			if changedNamespaces != nil {
				for _, namespace := range changedNamespaces {
					a.Options.Namespace = namespace
					err = a.applyNamespace()
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// ApplyAll is the entry point for performing a namespace apply on all namespaces.
// It checks if the working directory is in cloud-platform-environments, prepare the folder chunks based on the
// numRoutines the apply has to run, then loop over the list of namespaces in each chunk and does the kubectl apply
// and terraform init and apply of the namespace
func (a *Apply) ApplyAll() error {
	re := RepoEnvironment{}
	err := re.mustBeInCloudPlatformEnvironments()
	if err != nil {
		return err
	}

	repoPath := "namespaces/" + a.Options.ClusterCtx
	folderChunks, err := util.GetFolderChunks(repoPath, numRoutines)
	if err != nil {
		return err
	}

	for i := 0; i < len(folderChunks); i++ {
		err := a.applyNamespaceDirs(folderChunks[i])
		if err != nil {
			return err
		}

	}
	return nil
}

// applyNamespaceDirs get a folder chunk which is the list of namespaces, loop over each of them,
// get the latest changes (In case any PRs were merged since the pipeline started), and perform
// the apply of that namespace
func (a *Apply) applyNamespaceDirs(chunkFolder []string) error {
	for _, folder := range chunkFolder {

		// split the path to get the namespace name
		namespace := strings.Split(folder, "/")
		a.Options.Namespace = namespace[2]

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

// planKubectl calls the applier -> applyKubectl with dry-run enabled and return the output from applier
func (a *Apply) planKubectl() (string, error) {
	log.Printf("Running kubectl dry-run for namespace: %v in directory %v", a.Options.Namespace, a.Dir)

	outputKubectl, err := a.Applier.KubectlApply(a.Options.Namespace, a.Dir, true)
	if err != nil {
		err := fmt.Errorf("error running kubectl on namespace %s: in directory: %v, %v", a.Options.Namespace, a.Dir, err)
		return "", err
	}

	return outputKubectl, nil
}

// applyKubectl calls the applier -> applyKubectl with dry-run disabled and return the output from applier
func (a *Apply) applyKubectl() (string, error) {
	log.Printf("Running kubectl for namespace: %v in directory %v", a.Options.Namespace, a.Dir)

	outputKubectl, err := a.Applier.KubectlApply(a.Options.Namespace, a.Dir, false)
	if err != nil {
		err := fmt.Errorf("error running kubectl on namespace %s: %v", a.Options.Namespace, err)
		return "", err
	}

	return outputKubectl, nil
}

// planTerraform calls applier -> TerraformInitAndPlan and prints the output from applier
func (a *Apply) planTerraform() (string, error) {
	log.Printf("Running Terraform Plan for namespace: %v", a.Options.Namespace)

	tfFolder := a.Dir + "/resources"

	outputTerraform, err := a.Applier.TerraformInitAndPlan(a.Options.Namespace, tfFolder)
	if err != nil {
		err := fmt.Errorf("error running terraform on namespace %s: %v", a.Options.Namespace, err)
		return "", err
	}
	return outputTerraform, nil
}

// applyTerraform calls applier -> TerraformInitAndApply and prints the output from applier
func (a *Apply) applyTerraform() (string, error) {
	log.Printf("Running Terraform Apply for namespace: %v", a.Options.Namespace)

	tfFolder := a.Dir + "/resources"

	outputTerraform, err := a.Applier.TerraformInitAndApply(a.Options.Namespace, tfFolder)
	if err != nil {
		err := fmt.Errorf("error running terraform on namespace %s: %v", a.Options.Namespace, err)
		return "", err
	}
	return outputTerraform, nil
}

// planNamespace intiates a new Apply object with options and env variables, and calls the
// applyKubectl with dry-run enabled and calls applier TerraformInitAndPlan and prints the output
func (a *Apply) planNamespace() error {
	applier := NewApply(*a.Options)

	outputKubectl, err := applier.planKubectl()
	if err != nil {
		return err
	}

	outputTerraform, err := applier.planTerraform()
	if err != nil {
		return err
	}

	fmt.Println("\nOutput of kubectl:", outputKubectl)
	fmt.Println("\nOutput of terraform:")
	util.Redacted(os.Stdout, outputTerraform)
	return nil
}

// applyNamespace intiates a new Apply object with options and env variables, and calls the
// applyKubectl with dry-run disabled and calls applier TerraformInitAndApply and prints the output
func (a *Apply) applyNamespace() error {
	applier := NewApply(*a.Options)

	outputKubectl, err := applier.applyKubectl()
	if err != nil {
		return err
	}

	outputTerraform, err := applier.applyTerraform()
	if err != nil {
		return err
	}

	fmt.Println("\nOutput of kubectl:", outputKubectl)
	fmt.Println("\nOutput of terraform:")
	util.Redacted(os.Stdout, outputTerraform)
	return nil
}
