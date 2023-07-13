package environment

import (
	"fmt"
	"log"
	"os"
	"strings"

	gogithub "github.com/google/go-github/github"
	"github.com/kelseyhightower/envconfig"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/github"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/util"
)

// Options are used to configure plan/apply sessions.
// These options are normally passed via flags in a command line.
type Options struct {
	Namespace, KubecfgPath, ClusterCtx, GithubToken string
	PRNumber                                        int
	AllNamespaces                                   bool
	EnableApplySkip, RedactedEnv                    bool
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
	GithubClient    github.GithubIface
}

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
	a.RequiredEnvVars.clusterstatebucket = reqEnvVars.clusterstatebucket
	a.RequiredEnvVars.kubernetescluster = reqEnvVars.kubernetescluster
	a.RequiredEnvVars.githubowner = reqEnvVars.githubowner
	a.RequiredEnvVars.githubtoken = reqEnvVars.githubtoken
	a.RequiredEnvVars.pingdomapitoken = reqEnvVars.pingdomapitoken
	// Set KUBE_CONFIG_PATH to the path of the kubeconfig file
	// This is needed for terraform to be able to connect to the cluster when a different kubecfg is passed
	if err := os.Setenv("KUBE_CONFIG_PATH", a.Options.KubecfgPath); err != nil {
		log.Fatalln("KUBE_CONFIG_PATH environment variable cant be set:", err.Error())
	}
}

// Plan is the entry point for performing a namespace plan.
// It checks if the working directory is in cloud-platform-environments, checks if a PR number or a namespace is given
// If a namespace is given, it perform a `kubectl apply --dry-run=client` and a terraform init and plan of that namespace
// else checks for PR number and get the list of changed namespaces in the PR. Then does the `kubectl apply --dry-run=client` and
// terraform init and plan of all the namespaces changed in the PR
func (a *Apply) Plan() error {

	if a.Options.PRNumber == 0 && a.Options.Namespace == "" {
		return fmt.Errorf("either a PR Id/Number or a namespace is required to perform plan")
	}

	// If a namespace is given as a flag, then perform a plan for the given namespace.
	if a.Options.Namespace != "" {
		err := a.planNamespace()
		if err != nil {
			return err
		}
		return nil
	} else {
		files, err := a.GithubClient.GetChangedFiles(a.Options.PRNumber)
		if err != nil {
			return fmt.Errorf("failed to fetch list of changed files: %s in PR %v", err, a.Options.PRNumber)
		}
		changedNamespaces, err := nsChangedInPR(files, a.Options.ClusterCtx, false)
		if err != nil {
			return err
		}
		for _, namespace := range changedNamespaces {
			a.Options.Namespace = namespace
			err = a.planNamespace()
			if err != nil {
				return err
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
		err := fmt.Errorf("either a PR ID/Number or a namespace is required to perform apply")
		return err
	}
	// If a namespace is given as a flag, then perform a apply for the given namespace.
	if a.Options.Namespace != "" {
		err := a.applyNamespace()
		if err != nil {
			return err
		}
	} else {
		isMerged, err := a.GithubClient.IsMerged(a.Options.PRNumber)
		if err != nil {
			return err
		}
		if isMerged {
			repos, err := a.GithubClient.GetChangedFiles(a.Options.PRNumber)
			if err != nil {
				return err
			}

			changedNamespaces, err := nsChangedInPR(repos, a.Options.ClusterCtx, false)
			if err != nil {
				return err
			}
			for _, namespace := range changedNamespaces {
				a.Options.Namespace = namespace
				if _, err = os.Stat(a.Options.Namespace); err != nil {
					fmt.Println("Applying Namespace:", namespace)
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

// Destroy is the entry point for performing a namespace destroy.
// It checks if the working directory is in cloud-platform-environments, checks if a PR number is given and merged
// The method get the list of namespaces that are deleted in that merger PR, and for all namespaces in the PR does the
// terraform init and destroy and do a kubectl delete
func (a *Apply) Destroy() error {
	fmt.Println("Destroying Namespaces in PR", a.Options.PRNumber)
	if a.Options.PRNumber == 0 {
		err := fmt.Errorf("a PR ID/Number is required to perform destroy")
		return err
	}
	isMerged, err := a.GithubClient.IsMerged(a.Options.PRNumber)
	if err != nil {
		return err
	}
	if isMerged {
		changedNamespaces, err := a.nsCreateRawChangedFilesInPR(a.Options.ClusterCtx, a.Options.PRNumber)
		fmt.Println("Namespaces changed in PR", changedNamespaces)
		if err != nil {
			return err
		}
		for _, namespace := range changedNamespaces {
			a.Options.Namespace = namespace
			if _, err = os.Stat(a.Options.Namespace); err != nil {
				fmt.Println("Destroying Namespace:", namespace)
				err = a.destroyNamespace()
				if err != nil {
					return err
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
		err := fmt.Errorf("error running kubectl on namespace %s: in directory: %v, %v\n %v", a.Options.Namespace, a.Dir, err, outputKubectl)
		return "", err
	}

	return outputKubectl, nil
}

// applyKubectl calls the applier -> applyKubectl with dry-run disabled and return the output from applier
func (a *Apply) applyKubectl() (string, error) {
	log.Printf("Running kubectl for namespace: %v in directory %v", a.Options.Namespace, a.Dir)

	outputKubectl, err := a.Applier.KubectlApply(a.Options.Namespace, a.Dir, false)
	if err != nil {
		err := fmt.Errorf("error running kubectl on namespace %s: %v \n %v", a.Options.Namespace, err, outputKubectl)
		return "", err
	}

	return outputKubectl, nil
}

// deleteKubectl calls the applier -> deleteKubectl with dry-run disabled and return the output from applier
func (a *Apply) deleteKubectl() (string, error) {
	log.Printf("Running kubectl delete for namespace: %v in directory %v", a.Options.Namespace, a.Dir)

	outputKubectl, err := a.Applier.KubectlDelete(a.Options.Namespace, a.Dir, false)
	if err != nil {
		err := fmt.Errorf("error running kubectl delete on namespace %s: %v \n %v", a.Options.Namespace, err, outputKubectl)
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
		err := fmt.Errorf("error running terraform on namespace %s: %v \n %v", a.Options.Namespace, err, outputTerraform)
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
		err := fmt.Errorf("error running terraform on namespace %s: %v \n %v", a.Options.Namespace, err, outputTerraform)
		return "", err
	}
	return outputTerraform, nil
}

// applyTerraform calls applier -> TerraformInitAndDestroy and prints the output from applier
func (a *Apply) destroyTerraform() (string, error) {
	log.Printf("Running Terraform Destroy for namespace: %v", a.Options.Namespace)

	tfFolder := a.Dir + "/resources"

	outputTerraform, err := a.Applier.TerraformInitAndDestroy(a.Options.Namespace, tfFolder)
	if err != nil {
		err := fmt.Errorf("error running terraform on namespace %s: %v \n %v", a.Options.Namespace, err, outputTerraform)
		return "", err
	}
	return outputTerraform, nil
}

// planNamespace intiates a new Apply object with options and env variables, and calls the
// applyKubectl with dry-run enabled and calls applier TerraformInitAndPlan and prints the output
func (a *Apply) planNamespace() error {
	applier := NewApply(*a.Options)
	repoPath := "namespaces/" + a.Options.ClusterCtx + "/" + a.Options.Namespace

	if util.IsYamlFileExists(repoPath) {
		outputKubectl, err := applier.planKubectl()
		if err != nil {
			return err
		}

		fmt.Println("\nOutput of kubectl:", outputKubectl)
	} else {
		fmt.Printf("Namespace %s does not have yaml resources folder, skipping kubectl apply --dry-run\n", a.Options.Namespace)
	}

	exists, err := util.IsFilePathExists(repoPath + "/resources")
	if err == nil && exists {
		outputTerraform, err := applier.planTerraform()
		if err != nil {
			return err
		}

		fmt.Println("\nOutput of terraform:")
		util.RedactedEnv(os.Stdout, outputTerraform, a.Options.RedactedEnv)
	} else {
		fmt.Printf("Namespace %s does not have terraform resources folder, skipping terraform plan\n", a.Options.Namespace)
	}
	return nil
}

// secretBlockerExists takes a filepath (usually a namespace name i.e. namespaces/live.../mynamespace)
// and checks if the file SECRET_ROTATE_BLOCK exists.
func secretBlockerExists(filePath string) bool {
	// Check if the file contains a secret blocker
	// If it does, we don't want to apply it
	// If it doesn't, we do want to apply it
	secretBlocker := "SECRET_ROTATE_BLOCK"
	if _, err := os.Stat(filePath + "/" + secretBlocker); err == nil {
		return true
	}

	return false
}

// applySkipExists takes a filepath (usually a namespace name i.e. namespaces/live.../mynamespace)
// and checks if the file applySkipExists exists.
func applySkipExists(filePath string) bool {
	// Check if the file contains a apply skip, skip applying this namespace
	applySkip := "APPLY_PIPELINE_SKIP_THIS_NAMESPACE"
	if _, err := os.Stat(filePath + "/" + applySkip); err == nil {
		return true
	}

	return false
}

// applyNamespace intiates a new Apply object with options and env variables, and calls the
// applyKubectl with dry-run disabled and calls applier TerraformInitAndApply and prints the output
func (a *Apply) applyNamespace() error {
	// secretBlocker is a file used to control the behaviour of a namespace that will have all
	// secrets in a namespace rotated. This came out of the requirement to rotate IAM credentials
	// post circle breach.
	repoPath := "namespaces/" + a.Options.ClusterCtx + "/" + a.Options.Namespace

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		fmt.Printf("Namespace %s does not exist, skipping apply\n", a.Options.Namespace)
		return nil
	}

	if secretBlockerExists(repoPath) {
		log.Printf("Namespace %s has a secret rotation blocker file, skipping apply", a.Options.Namespace)
		// We don't want to return an error here so we softly fail.
		return nil
	}

	if (a.Options.EnableApplySkip) && (applySkipExists(repoPath)) {
		log.Printf("Namespace %s has a apply skip file, skipping apply", a.Options.Namespace)
		// We don't want to return an error here so we softly fail.
		return nil
	}

	applier := NewApply(*a.Options)

	if util.IsYamlFileExists(repoPath) {
		outputKubectl, err := applier.applyKubectl()
		if err != nil {
			return err
		}

		fmt.Println("\nOutput of kubectl:", outputKubectl)
	} else {
		fmt.Printf("Namespace %s does not have yaml resources folder, skipping kubectl apply", a.Options.Namespace)
	}

	exists, err := util.IsFilePathExists(repoPath + "/resources")
	// Set KUBE_CONFIG_PATH to the path of the kubeconfig file
	// This is needed for terraform to be able to connect to the cluster when a different kubecfg is passed
	if err := os.Setenv("KUBE_CONFIG_PATH", a.Options.KubecfgPath); err != nil {
		return err
	}
	if err == nil && exists {
		outputTerraform, err := applier.applyTerraform()
		if err != nil {
			return err
		}

		fmt.Println("\nOutput of terraform:")
		util.RedactedEnv(os.Stdout, outputTerraform, a.Options.RedactedEnv)
	} else {
		fmt.Printf("Namespace %s does not have terraform resources folder, skipping terraform apply", a.Options.Namespace)
	}
	return nil
}

// destroyNamespace intiates a apply object with options and env variables, and calls the
// calls applier TerraformInitAndDestroy, applyKubectl with dry-run disabled and prints the output
func (a *Apply) destroyNamespace() error {
	repoPath := "namespaces/" + a.Options.ClusterCtx + "/" + a.Options.Namespace

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		fmt.Printf("Namespace %s does not exist, skipping destroy\n", a.Options.Namespace)
		return nil
	}

	applier := NewApply(*a.Options)

	exists, err := util.IsFilePathExists(repoPath + "/resources")
	if err == nil && exists {
		outputTerraform, err := applier.destroyTerraform()
		if err != nil {
			return err
		}

		fmt.Println("\nOutput of terraform:")
		util.RedactedEnv(os.Stdout, outputTerraform, a.Options.RedactedEnv)

		if util.IsYamlFileExists(repoPath) {
			outputKubectl, err := applier.deleteKubectl()
			if err != nil {
				return err
			}

			fmt.Println("\nOutput of kubectl:", outputKubectl)
		} else {
			fmt.Printf("Namespace %s does not have yaml resources folder, skipping kubectl delete", a.Options.Namespace)
		}

	} else {
		fmt.Printf("Namespace %s does not have terraform resources folder, skipping terraform destroy", a.Options.Namespace)
	}
	return nil
}

// nsCreateRawChangedFilesInPR get the list of changed files for a given PR. checks if the file is deleted and
// write the deleted file to the namespace folder
func (a *Apply) nsCreateRawChangedFilesInPR(cluster string, prNumber int) ([]string, error) {
	files, err := a.GithubClient.GetChangedFiles(prNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch list of changed files: %s", err)
	}

	namespaces, err := nsChangedInPR(files, a.Options.ClusterCtx, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace for destroy from the PR: %s", err)
	}
	if len(namespaces) == 0 {
		fmt.Println("No namespace found in the PR for destroy")
		return nil, nil
	}
	err = createNamespaceforDestroy(namespaces, cluster)
	if err != nil {
		return nil, fmt.Errorf("failed to create namespace for destroy: %s", err)
	}

	// Get the contents of the CommitFile from RawURL
	// https://developer.github.com/v3/repos/contents/#get-contents

	for _, file := range files {
		data, err := util.GetGithubRawContents(file.GetRawURL())
		if err != nil {
			return nil, fmt.Errorf("failed to get raw contents: %s", err)
		}
		// Create List with changed files
		if err := os.WriteFile(*file.Filename, data, 0644); err != nil {
			return nil, fmt.Errorf("failed to write file list: %s", err)
		}
	}
	return namespaces, nil

}

// nsChangedInPR get the list of changed files for a given PR. checks if the namespaces exists in the given cluster
// folder and return the list of namespaces.
func nsChangedInPR(files []*gogithub.CommitFile, cluster string, isDeleted bool) ([]string, error) {
	var namespaceNames []string
	for _, file := range files {
		// check of the file is a deleted file
		if isDeleted && *file.Status != "removed" {
			fmt.Println("some of files are not marked for deletion: file", *file.Filename, "is not deleted")
			return nil, nil
		}

		// namespaces filepaths are assumed to come in
		// the format: namespaces/<cluster>.cloud-platform.service.justice.gov.uk/<namespaceName>
		s := strings.Split(*file.Filename, "/")
		//only get namespaces from the folder that belong to the given cluster and
		// ignore changes outside namespace directories
		if len(s) > 1 && s[1] == cluster {
			namespaceNames = append(namespaceNames, s[2])
		}
	}
	return util.DeduplicateList(namespaceNames), nil
}

func createNamespaceforDestroy(namespaces []string, cluster string) error {
	wd, _ := os.Getwd()
	for _, ns := range namespaces {
		// make directory if it doesn't exist
		if _, err := os.Stat(wd + "/namespaces/" + cluster + "/" + ns); err != nil {
			err := os.Mkdir(wd+"/namespaces/"+cluster+"/"+ns, 0755)
			if err != nil {
				return fmt.Errorf("error creating namespaces directory: %s", err)
			}
			err = os.Mkdir(wd+"/namespaces/"+cluster+"/"+ns+"/resources", 0755)
			if err != nil {
				return fmt.Errorf("error creating resources directory: %s", err)
			}
		} else {
			return fmt.Errorf("error creating directory, namespace exists in the environments repo: %s", err)
		}
	}
	return nil
}
