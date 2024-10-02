package environment

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/github"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/slack"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/util"
)

// Options are used to configure plan/apply sessions.
// These options are normally passed via flags in a command line.
type Options struct {
	Namespace, KubecfgPath, ClusterCtx, ClusterDir, GithubToken string
	PRNumber                                                    int
	BuildUrl                                                    string
	AllNamespaces                                               bool
	EnableApplySkip, RedactedEnv, SkipProdDestroy               bool
	BatchApplyIndex, BatchApplySize                             int
	OnlySkipFileChanged, IsApplyPipeline                        bool
}

// RequiredEnvVars is used to store values such as TF_VAR_ , github and pingdom tokens
// which are needed to perform terraform operations for a given namespace
type RequiredEnvVars struct {
	clustername        string `required:"true" envconfig:"TF_VAR_cluster_name"`
	clusterstatebucket string `required:"true" envconfig:"TF_VAR_cluster_state_bucket"`
	kubernetescluster  string `required:"true" envconfig:"TF_VAR_kubernetes_cluster"`
	githubowner        string `required:"true" envconfig:"TF_VAR_github_owner"`
	githubtoken        string `required:"true" envconfig:"TF_VAR_github_token"`
	SlackBotToken      string `required:"false" envconfig:"SLACK_BOT_TOKEN"`
	SlackWebhookUrl    string `required:"false" envconfig:"SLACK_WEBHOOK_URL"`
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

func notifyUserApplyFailed(prNumberInt int, slackToken, webhookUrl, buildUrl string) {
	if prNumberInt > 0 && strings.Contains(buildUrl, "http") {
		prNumber := fmt.Sprintf("%d", prNumberInt)

		slackErr := slack.Notify(prNumber, slackToken, webhookUrl, buildUrl)

		if slackErr != nil {
			fmt.Printf("Warning: Error notifying user of build error %v\n", slackErr)
		}
	}
}

// NewApply creates a new Apply object and populates its fields with values from options(which are flags),
// instantiate Applier object which also checks and sets the Backend config variables to do terraform init,
// RequiredEnvVars object which stores the values required for plan/apply of namespace
func NewApply(opt Options) *Apply {
	apply := Apply{
		Options: &opt,
		Applier: NewApplier("/usr/local/bin/terraform", "/usr/local/bin/kubectl"),
		Dir:     "namespaces/" + opt.ClusterDir + "/" + opt.Namespace,
	}

	apply.Initialize()
	return &apply
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

			a.Options.OnlySkipFileChanged = false

			if len(repos) == 1 {
				a.Options.OnlySkipFileChanged = strings.Contains(*repos[0].Filename, "APPLY_PIPELINE_SKIP_THIS_NAMESPACE")
			}

			if err != nil {
				return err
			}

			changedNamespaces, err := nsChangedInPR(repos, a.Options.ClusterDir, false)
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

// ApplyAll is the entry point for performing a namespace apply on all namespaces.
// It checks if the working directory is in cloud-platform-environments, get the list of namespace folders and perform kubectl apply
// and terraform init and apply of the namespace
func (a *Apply) ApplyAll() error {
	re := RepoEnvironment{}
	err := re.mustBeInCloudPlatformEnvironments()
	if err != nil {
		return err
	}

	repoPath := "namespaces/" + a.Options.ClusterDir
	folders, err := util.ListFolderPaths(repoPath)
	if err != nil {
		return err
	}

	// skip the root folder namespaces/cluster.cloud-platform.service.justice.gov.uk which is the first
	// element of the slice. We dont want to apply from the root folder
	var nsFolders []string
	nsFolders = append(nsFolders, folders[1:]...)

	err = a.applyNamespaceDirs(nsFolders)
	if err != nil {
		return err
	}

	return nil
}

// ApplyBatch is the entry point for performing a namespace apply on a batch of namespaces.
// It checks if the working directory is in cloud-platform-environments, get the list of namespace folders based on the batch index and size
// and perform kubectl apply and terraform init and apply of the namespace
func (a *Apply) ApplyBatch() error {
	re := RepoEnvironment{}
	err := re.mustBeInCloudPlatformEnvironments()
	if err != nil {
		return err
	}

	repoPath := "namespaces/" + a.Options.ClusterDir
	folderChunks, err := util.GetFolderChunks(repoPath, a.Options.BatchApplyIndex, a.Options.BatchApplySize)
	if err != nil {
		return err
	}

	err = a.applyNamespaceDirs(folderChunks)
	if err != nil {
		return err
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

// applyTerraform calls applier -> TerraformInitAndApply and prints the output from applier
func (a *Apply) applyTerraform() (string, error) {
	log.Printf("Running Terraform Apply for namespace: %v", a.Options.Namespace)

	tfFolder := a.Dir + "/resources"

	outputTerraform, err := a.Applier.TerraformInitAndApply(a.Options.Namespace, tfFolder)

	if a.Options.IsApplyPipeline && err != nil {
		versionDescription, filenames, updateErr := checkRdsAndUpdate(err.Error(), tfFolder)

		if updateErr != nil {
			return "", fmt.Errorf("update error running terraform on namespace %s: %s \n %s \n %s", a.Options.Namespace, err.Error(), updateErr.Error(), outputTerraform)
		}

		description := "\n\n``` " + versionDescription + " ```\n\n" + a.Options.BuildUrl

		prUrl, createErr := createPR(description, a.Options.Namespace, a.Options.GithubToken, "cloud-platform-environments")(a.GithubClient, filenames)

		if createErr != nil {
			return "", fmt.Errorf("create error running terraform on namespace %s: %v \n %v \n %v", a.Options.Namespace, err, outputTerraform, createErr)
		}

		postPR(prUrl, a.RequiredEnvVars.SlackWebhookUrl)

	}
	if err != nil {
		return "", fmt.Errorf("error running terraform on namespace %s: %v \n %v", a.Options.Namespace, err, outputTerraform)
	}
	return outputTerraform, nil
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
	repoPath := "namespaces/" + a.Options.ClusterDir + "/" + a.Options.Namespace

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
			if !a.Options.OnlySkipFileChanged && !a.Options.IsApplyPipeline {
				notifyUserApplyFailed(a.Options.PRNumber, applier.RequiredEnvVars.SlackBotToken, applier.RequiredEnvVars.SlackWebhookUrl, a.Options.BuildUrl)
			}
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
		applier.GithubClient = a.GithubClient
		outputTerraform, err := applier.applyTerraform()
		if err != nil {
			if !a.Options.OnlySkipFileChanged && !a.Options.IsApplyPipeline {
				notifyUserApplyFailed(a.Options.PRNumber, applier.RequiredEnvVars.SlackBotToken, applier.RequiredEnvVars.SlackWebhookUrl, a.Options.BuildUrl)
			}
			return err
		}
		fmt.Println("\nOutput of terraform:")
		util.RedactedEnv(os.Stdout, outputTerraform, a.Options.RedactedEnv)
	} else {
		fmt.Printf("Namespace %s does not have terraform resources folder, skipping terraform apply", a.Options.Namespace)
	}
	return nil
}
