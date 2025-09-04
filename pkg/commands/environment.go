package commands

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	environment "github.com/ministryofjustice/cloud-platform-cli/pkg/environment"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/github"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/util/homedir"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

// variables specific to commands package used to store the values of flags of various environment sub commands
var (
	module, moduleVersion string
	optFlags              environment.Options
)

// skipEnvCheck is a flag to skip the environments repository check.
// This is useful for testing.
var skipEnvCheck bool

// answersFile is a flag to specify the path to the answers file.
var answersFile string

var clusterName, githubToken string

func addEnvironmentCmd(topLevel *cobra.Command) {
	topLevel.AddCommand(environmentCmd)
	envSubCommands := []*cobra.Command{
		environmentApplyCmd,
		environmentBumpModuleCmd,
		environmentCreateCmd,
		environmentDestroyCmd,
		environmentDivergenceCmd,
		environmentEcrCmd,
		environmentPlanCmd,
		environmentPrototypeCmd,
		environmentRdsCmd,
		environmentS3Cmd,
		environmentSvcCmd,
		environmentRdsDriftCheckerCmd,
	}

	for _, cmd := range envSubCommands {
		environmentCmd.AddCommand(cmd)
	}

	environmentEcrCmd.AddCommand(environmentEcrCreateCmd)
	environmentRdsCmd.AddCommand(environmentRdsCreateCmd)
	environmentS3Cmd.AddCommand(environmentS3CreateCmd)
	environmentSvcCmd.AddCommand(environmentSvcCreateCmd)
	environmentPrototypeCmd.AddCommand(environmentPrototypeCreateCmd)

	// flags
	environmentApplyCmd.Flags().BoolVar(&optFlags.AllNamespaces, "all-namespaces", false, "Apply all namespaces with -all-namespaces")
	environmentApplyCmd.Flags().IntVar(&optFlags.BatchApplyIndex, "batch-apply-index", 0, "Starting index for Apply to a batch of namespaces")
	environmentApplyCmd.Flags().IntVar(&optFlags.BatchApplySize, "batch-apply-size", 0, "Number of namespaces to apply in a batch")
	environmentApplyCmd.Flags().BoolVar(&optFlags.EnableApplySkip, "enable-apply-skip", false, "Enable skipping apply for a namespace")
	environmentApplyCmd.Flags().StringVarP(&optFlags.Namespace, "namespace", "n", "", "Namespace which you want to perform the apply")
	environmentApplyCmd.Flags().IntVar(&optFlags.PRNumber, "pr-number", 0, "Pull request ID or number to which you want to perform the apply")
	// Re-use the environmental variable TF_VAR_github_token to call Github Client which is needed to perform terraform operations on each namespace
	environmentApplyCmd.Flags().StringVar(&optFlags.GithubToken, "github-token", os.Getenv("TF_VAR_github_token"), "Personal access Token from Github ")
	environmentApplyCmd.Flags().StringVar(&optFlags.KubecfgPath, "kubecfg", filepath.Join(homedir.HomeDir(), ".kube", "config"), "path to kubeconfig file")
	environmentApplyCmd.Flags().StringVar(&optFlags.ClusterCtx, "cluster", "", "cluster context from kubeconfig file")
	environmentApplyCmd.Flags().StringVar(&optFlags.ClusterDir, "clusterdir", "", "folder name under namespaces/ inside cloud-platform-environments repo referring to full cluster name")
	environmentApplyCmd.PersistentFlags().BoolVar(&optFlags.RedactedEnv, "redact", true, "Redact the terraform output before printing")
	environmentApplyCmd.Flags().StringVar(&optFlags.BuildUrl, "build-url", "", "The concourse apply build url")
	environmentApplyCmd.Flags().BoolVar(&optFlags.IsApplyPipeline, "is-apply-pipeline", false, "is this running in the apply pipelines")

	environmentBumpModuleCmd.Flags().StringVarP(&module, "module", "m", "", "Module to upgrade the version")
	environmentBumpModuleCmd.Flags().StringVarP(&moduleVersion, "module-version", "v", "", "Semantic version to bump a module to")

	environmentCreateCmd.Flags().BoolVarP(&skipEnvCheck, "skip-env-check", "s", false, "Skip the environment check")
	environmentCreateCmd.Flags().StringVarP(&answersFile, "answers-file", "a", "", "Path to the answers file")

	// e.g. if this is the Pull request to perform the apply: https://github.com/ministryofjustice/cloud-platform-environments/pull/8370, the pr ID is 8370.
	environmentDestroyCmd.Flags().IntVar(&optFlags.PRNumber, "pr-number", 0, "Pull request ID or number to which you want to perform the destroy")
	environmentDestroyCmd.Flags().StringVarP(&optFlags.Namespace, "namespace", "n", "", "Namespace which you want to perform the destroy")

	// Re-use the environmental variable TF_VAR_github_token to call Github Client which is needed to perform terraform operations on each namespace
	environmentDestroyCmd.Flags().StringVar(&optFlags.GithubToken, "github-token", os.Getenv("TF_VAR_github_token"), "Personal access Token from Github ")
	environmentDestroyCmd.Flags().StringVar(&optFlags.KubecfgPath, "kubecfg", filepath.Join(homedir.HomeDir(), ".kube", "config"), "path to kubeconfig file")
	environmentDestroyCmd.Flags().StringVar(&optFlags.ClusterCtx, "cluster", "", "cluster context from kubeconfig file")
	environmentDestroyCmd.Flags().StringVar(&optFlags.ClusterDir, "clusterdir", "", "folder name under namespaces/ inside cloud-platform-environments repo referring to full cluster name")
	environmentDestroyCmd.PersistentFlags().BoolVar(&optFlags.RedactedEnv, "redact", true, "Redact the terraform output before printing")
	environmentDestroyCmd.Flags().BoolVar(&optFlags.SkipProdDestroy, "skip-prod-destroy", true, "skip prod namespaces from destroy namespace")

	environmentDivergenceCmd.Flags().StringVarP(&clusterName, "cluster-name", "c", "live", "[optional] Cluster name")
	environmentDivergenceCmd.Flags().StringVarP(&githubToken, "github-token", "g", "", "[required] Github token")
	environmentDivergenceCmd.Flags().StringVarP(&kubeconfig, "kubeconfig", "k", "", "[optional] Kubeconfig file path")
	if err := environmentDivergenceCmd.MarkFlagRequired("github-token"); err != nil {
		log.Fatal(err)
	}

	// e.g. if this is the Pull request to perform the apply: https://github.com/ministryofjustice/cloud-platform-environments/pull/8370, the pr ID is 8370.
	environmentPlanCmd.Flags().IntVar(&optFlags.PRNumber, "pr-number", 0, "Pull request ID or number to which you want to perform the plan")
	environmentPlanCmd.Flags().StringVarP(&optFlags.Namespace, "namespace", "n", "", "Namespace which you want to perform the plan")
	// Re-use the environmental variable TF_VAR_github_token to call Github Client which is needed to perform terraform operations on each namespace
	environmentPlanCmd.Flags().StringVar(&optFlags.GithubToken, "github-token", os.Getenv("TF_VAR_github_token"), "Personal access Token from Github ")
	environmentPlanCmd.Flags().StringVar(&optFlags.KubecfgPath, "kubecfg", filepath.Join(homedir.HomeDir(), ".kube", "config"), "path to kubeconfig file")
	environmentPlanCmd.Flags().StringVar(&optFlags.ClusterCtx, "cluster", "", "cluster context from kubeconfig file")
	environmentPlanCmd.Flags().StringVar(&optFlags.ClusterDir, "clusterdir", "", "folder name under namespaces/ inside cloud-platform-environments repo referring to full cluster name")
	environmentPlanCmd.PersistentFlags().BoolVar(&optFlags.RedactedEnv, "redact", true, "Redact the terraform output before printing")
	environmentPlanCmd.Flags().StringVar(&optFlags.AppID, "github-appid", os.Getenv("TF_VAR_github_cloud_platform_concourse_bot_app_id"), "App ID ")
	environmentPlanCmd.Flags().StringVar(&optFlags.InstallID, "github-installation-id", os.Getenv("TF_VAR_github_cloud_platform_concourse_bot_installation_id"), "Installation ID ")
	environmentPlanCmd.Flags().StringVar(&optFlags.PemFile, "github-pem-file", os.Getenv("TF_VAR_github_cloud_platform_concourse_bot_pem_file"), "PEM file ")

}

var environmentCmd = &cobra.Command{
	Use:    "environment",
	Short:  `Cloud Platform Environment actions`,
	PreRun: upgradeIfNotLatest,
}

var environmentCreateCmd = &cobra.Command{
	Use:   "create",
	Short: `Create an environment`,
	Example: heredoc.Doc(`
	> cloud-platform environment create
	`),
	PreRun: upgradeIfNotLatest,
	RunE: func(cmd *cobra.Command, args []string) error {
		return environment.CreateTemplateNamespace(skipEnvCheck, answersFile)
	},
}

var environmentEcrCmd = &cobra.Command{
	Use:   "ecr",
	Short: `Add an ECR to a namespace`,
	Example: heredoc.Doc(`
	> cloud-platform environment ecr create
	`),
	PreRun: upgradeIfNotLatest,
}

var environmentPlanCmd = &cobra.Command{
	Use: "plan",
	Short: `Perform a terraform plan and kubectl apply --dry-run=client for a given namespace using either -namespace flag or the
	the namespace in the given PR Id/Number`,
	Long: `
	Perform a kubectl apply --dry-run=client and a terraform plan for a given namespace using either -namespace flag or the
	the namespace in the given PR Id/Number

	Along with the mandatory input flag, the below environments variables needs to be set
	TF_VAR_cluster_name - e.g. "cp-1902-02" to get the vpc details for some modules like rds, es
	TF_VAR_cluster_state_bucket - State where the cluster state is stored
	TF_VAR_cluster_state_key - folder name/state key inside the state bucket where cluster state is stored
	TF_VAR_github_owner - Github owner: ministryofjustice
	TF_VAR_github_token - Personal access token with repo scope to push github action secrets
	TF_VAR_kubernetes_cluster - Full name of the Cluster e.g. XXXXXX.gr7.eu-west2.eks.amazonaws.com
	PINGDOM_API_TOKEN - API Token to access pingdom
	PIPELINE_TERRAFORM_STATE_LOCK_TABLE - DynamoDB table where the state lock is stored
	PIPELINE_STATE_BUCKET - State bucket where the environments state is stored e.g cloud-platform-terraform-state
	PIPELINE_STATE_KEY_PREFIX - State key/ folder where the environments terraform state is stored e.g cloud-platform-environments
	PIPELINE_STATE_REGION - State region of the bucket e.g. eu-west-1
	PIPELINE_CLUSTER - Cluster name/folder inside namespaces/ in cloud-platform-environments
	PIPELINE_CLUSTER_STATE - Cluster name/folder inside the state bucket where the environments terraform state is stored. for "live" the state is stored under "live-1.cloud-platform.service..."
	`,
	Example: heredoc.Doc(`
	$ cloud-platform environment plan
	`),
	PreRun: upgradeIfNotLatest,
	Run: func(cmd *cobra.Command, args []string) {

		contextLogger := log.WithFields(log.Fields{"subcommand": "plan"})

		ghConfig := &github.GithubClientConfig{
			Repository: "cloud-platform-environments",
			Owner:      "ministryofjustice",
		}

		// get authtype this is only needed for migration purposes once users are all using github app this can be removed
		authType, err := github.NewGithubAppClient(ghConfig, optFlags.PemFile, optFlags.AppID, optFlags.InstallID).SearchAuthTypeDefaultInPR(context.Background(), optFlags.PRNumber)
		if err != nil {
			contextLogger.Fatalf("Failed to get auth_type from PR: %v", err)
		}

		var applier *environment.Apply
		if authType == "app" {
			applier = &environment.Apply{
				Options:      &optFlags,
				GithubClient: github.NewGithubAppClient(ghConfig, optFlags.PemFile, optFlags.AppID, optFlags.InstallID),
			}
		} else {
			applier = &environment.Apply{
				Options:      &optFlags,
				GithubClient: github.NewGithubClient(ghConfig, optFlags.GithubToken),
			}
		}

		err = applier.Plan()
		if err != nil {
			contextLogger.Fatal(err)
		}
	},
}

var environmentApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: `Perform a terraform apply and kubectl apply for a given namespace`,
	Long: `
	Perform a kubectl apply and a terraform apply for a given namespace using either -namespace flag or the
	the namespace in the given PR Id/Number

	Along with the mandatory input flag, the below environments variables needs to be set
	TF_VAR_cluster_name - e.g. "cp-1902-02" to get the vpc details for some modules like rds, es
	TF_VAR_cluster_state_bucket - State where the cluster state is stored
	TF_VAR_cluster_state_key - folder name/state key inside the state bucket where cluster state is stored
	TF_VAR_github_owner - Github owner: ministryofjustice
	TF_VAR_github_token - Personal access token with repo scope to push github action secrets
	TF_VAR_kubernetes_cluster - Full name of the Cluster e.g. XXXXXX.gr7.eu-west2.eks.amazonaws.com
	PINGDOM_API_TOKEN - API Token to access pingdom
	PIPELINE_TERRAFORM_STATE_LOCK_TABLE - DynamoDB table where the state lock is stored
	PIPELINE_STATE_BUCKET - State bucket where the environments state is stored e.g cloud-platform-terraform-state
	PIPELINE_STATE_KEY_PREFIX - State key/ folder where the environments terraform state is stored e.g cloud-platform-environments
	PIPELINE_STATE_REGION - State region of the bucket e.g. eu-west-1
	PIPELINE_CLUSTER - Cluster name/folder inside namespaces/ in cloud-platform-environments
	PIPELINE_CLUSTER_STATE - Cluster name/folder inside the state bucket where the environments terraform state is stored
	`,
	Example: heredoc.Doc(`
	$ cloud-platform environment apply -n <namespace>
	`),
	PreRun: upgradeIfNotLatest,
	Run: func(cmd *cobra.Command, args []string) {
		contextLogger := log.WithFields(log.Fields{"subcommand": "apply"})

		ghConfig := &github.GithubClientConfig{
			Repository: "cloud-platform-environments",
			Owner:      "ministryofjustice",
		}

		// get authtype this is only needed for migration purposes once users are all using github app this can be removed
		authType, err := github.NewGithubAppClient(ghConfig, optFlags.PemFile, optFlags.AppID, optFlags.InstallID).SearchAuthTypeDefaultInPR(context.Background(), optFlags.PRNumber)
		if err != nil {
			contextLogger.Fatalf("Failed to get auth_type from PR: %v", err)
		}

		var applier *environment.Apply
		if authType == "app" {
			applier = &environment.Apply{
				Options:      &optFlags,
				GithubClient: github.NewGithubAppClient(ghConfig, optFlags.PemFile, optFlags.AppID, optFlags.InstallID),
			}
		} else {
			applier = &environment.Apply{
				Options:      &optFlags,
				GithubClient: github.NewGithubClient(ghConfig, optFlags.GithubToken),
			}
		}

		// if -namespace or a prNumber is provided, apply on given namespace
		if optFlags.Namespace != "" || optFlags.PRNumber > 0 {
			err := applier.Apply()
			if err != nil {
				contextLogger.Fatal(err)
			}
			return
		}
		// if -batch-apply-index and -batch-apply-size is provided, apply on given batch of namespaces
		if optFlags.BatchApplyIndex >= 0 && optFlags.BatchApplySize > 0 {
			err := applier.ApplyBatch()
			if err != nil {
				contextLogger.Fatal(err)
			}
			return
		}
		// if -all-namespaces is provided, apply all namespaces
		if optFlags.AllNamespaces {
			err := applier.ApplyAll()
			if err != nil {
				contextLogger.Fatal(err)
			}
			return
		}
	},
}

var environmentDestroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: `Perform a terraform destroy and kubectl delete for a given namespace`,
	Long: `
	Perform a kubectl destroy and a terraform delete for a given namespace using either -namespace flag or the
	the namespace in the given PR Id/Number

	Along with the mandatory input flag, the below environments variables needs to be set
	TF_VAR_cluster_name - e.g. "cp-1902-02" to get the vpc details for some modules like rds, es
	TF_VAR_cluster_state_bucket - State where the cluster state is stored
	TF_VAR_cluster_state_key - folder name/state key inside the state bucket where cluster state is stored
	TF_VAR_github_owner - Github owner: ministryofjustice
	TF_VAR_github_token - Personal access token with repo scope to push github action secrets
	TF_VAR_kubernetes_cluster - Full name of the Cluster e.g. XXXXXX.gr7.eu-west2.eks.amazonaws.com
	PINGDOM_API_TOKEN - API Token to access pingdom
	PIPELINE_TERRAFORM_STATE_LOCK_TABLE - DynamoDB table where the state lock is stored
	PIPELINE_STATE_BUCKET - State bucket where the environments state is stored e.g cloud-platform-terraform-state
	PIPELINE_STATE_KEY_PREFIX - State key/ folder where the environments terraform state is stored e.g cloud-platform-environments
	PIPELINE_STATE_REGION - State region of the bucket e.g. eu-west-1
	PIPELINE_CLUSTER - Cluster name/folder inside namespaces/ in cloud-platform-environments
	PIPELINE_CLUSTER_STATE - Cluster name/folder inside the state bucket where the environments terraform state is stored
	`,
	Example: heredoc.Doc(`
	$ cloud-platform environment destroy -n <namespace>
	`),
	PreRun: upgradeIfNotLatest,
	Run: func(cmd *cobra.Command, args []string) {
		contextLogger := log.WithFields(log.Fields{"subcommand": "destroy"})

		ghConfig := &github.GithubClientConfig{
			Repository: "cloud-platform-environments",
			Owner:      "ministryofjustice",
		}

		// get authtype this is only needed for migration purposes once users are all using github app this can be removed
		authType, err := github.NewGithubAppClient(ghConfig, optFlags.PemFile, optFlags.AppID, optFlags.InstallID).SearchAuthTypeDefaultInPR(context.Background(), optFlags.PRNumber)
		if err != nil {
			contextLogger.Fatalf("Failed to get auth_type from PR: %v", err)
		}

		var applier *environment.Apply
		if authType == "app" {
			applier = &environment.Apply{
				Options:      &optFlags,
				GithubClient: github.NewGithubAppClient(ghConfig, optFlags.PemFile, optFlags.AppID, optFlags.InstallID),
			}
		} else {
			applier = &environment.Apply{
				Options:      &optFlags,
				GithubClient: github.NewGithubClient(ghConfig, optFlags.GithubToken),
			}
		}

		err = applier.Destroy()
		if err != nil {
			contextLogger.Fatal(err)
		}
	},
}

var environmentEcrCreateCmd = &cobra.Command{
	Use:    "create",
	Short:  `Create "resources/ecr.tf" terraform file for an ECR`,
	PreRun: upgradeIfNotLatest,
	RunE:   environment.CreateTemplateEcr,
}

var environmentRdsCmd = &cobra.Command{
	Use:   "rds",
	Short: `Add an RDS instance to a namespace`,
	Example: heredoc.Doc(`
	> cloud-platform environment rds create
	`),
	PreRun: upgradeIfNotLatest,
}

var environmentRdsCreateCmd = &cobra.Command{
	Use:    "create",
	Short:  `Create "resources/rds.tf" terraform file for an RDS instance`,
	PreRun: upgradeIfNotLatest,
	RunE:   environment.CreateTemplateRds,
}

var environmentS3Cmd = &cobra.Command{
	Use:   "s3",
	Short: `Add a S3 bucket to a namespace`,
	Example: heredoc.Doc(`
	> cloud-platform environment s3 create
	`),
	PreRun: upgradeIfNotLatest,
}

var environmentS3CreateCmd = &cobra.Command{
	Use:    "create",
	Short:  `Create "resources/s3.tf" terraform file for a S3 bucket`,
	PreRun: upgradeIfNotLatest,
	RunE:   environment.CreateTemplateS3,
}

var environmentSvcCmd = &cobra.Command{
	Use:   "serviceaccount",
	Short: `Add a serviceaccount to a namespace`,
	Example: heredoc.Doc(`
	> cloud-platform environment serviceaccount
	`),
	PreRun: upgradeIfNotLatest,
}

var environmentSvcCreateCmd = &cobra.Command{
	Use:   "create",
	Short: `Creates a serviceaccount in your chosen namespace`,
	Example: heredoc.Doc(`
	> cloud-platform environment serviceaccount create
	`),
	PreRun: upgradeIfNotLatest,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := environment.CreateTemplateServiceAccount(); err != nil {
			return err
		}

		return nil
	},
}

var environmentPrototypeCmd = &cobra.Command{
	Use:   "prototype",
	Short: `Create a gov.uk prototype kit site on the cloud platform`,
	Example: heredoc.Doc(`
	> cloud-platform environment prototype
	`),
	PreRun: upgradeIfNotLatest,
}

var environmentPrototypeCreateCmd = &cobra.Command{
	Use:   "create",
	Short: `Create an environment to host gov.uk prototype kit site on the cloud platform`,
	Long: `
Create a namespace folder and files in an existing prototype github repository to host a Gov.UK
Prototype Kit website on the Cloud Platform.

The namespace name should be your prototype github repository name:

  https://github.com/ministryofjustice/[repository name]
	`,
	Example: heredoc.Doc(`
	> cloud-platform environment prototype create
	`),
	PreRun: upgradeIfNotLatest,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := environment.CreateTemplatePrototype(); err != nil {
			return err
		}

		return nil
	},
}

var environmentBumpModuleCmd = &cobra.Command{
	Use:   "bump-module",
	Short: `Bump all specified module versions`,
	Example: heredoc.Doc(`
cloud-platform environments bump-module --module serviceaccount --module-version 1.1.1

Would bump all users serviceaccount modules in the environments repository to the specified version.
	`),
	PreRun: upgradeIfNotLatest,
	RunE: func(cmd *cobra.Command, args []string) error {
		if moduleVersion == "" || module == "" {
			return errors.New("--module and --module-version are required")
		}

		if err := environment.BumpModule(module, moduleVersion); err != nil {
			return err
		}
		return nil
	},
}

var environmentDivergenceCmd = &cobra.Command{
	Use:   "divergence",
	Short: `Check for divergence between the environments repository and the cluster`,
	Example: heredoc.Doc(`
	> cloud-platform environment divergence --cluster myTestCluster --githubToken myGithubToken123
	`),
	PreRun: upgradeIfNotLatest,
	Run: func(cmd *cobra.Command, args []string) {
		contextLogger := log.WithFields(log.Fields{"subcommand": "divergence"})
		// list of excluded Kubernetes namespaces to check.
		excludedNamespaces := []string{
			"cert-manager",
			"default",
			"ingress-controllers",
			"kube-node-lease",
			"kube-public",
			"kube-system",
			"kuberos",
			"logging",
			"gatekeeper-system",
			"overprovision",
			"velero",
			"trivy-system",
		}

		divergence, err := environment.NewDivergence(clusterName, kubeconfig, githubToken, excludedNamespaces)
		if err != nil {
			contextLogger.Fatal(err)
		}

		if err := divergence.Check(); err != nil {
			contextLogger.Fatal(err)
		}
	},
}

var environmentRdsDriftCheckerCmd = &cobra.Command{
	Use:   "rds-drift-checker <file-location>",
	Short: "Detect and correct RDS engine version drift from a CSV file in S3 or locally",
	Example: heredoc.Doc(`
		Run with a file from S3:
		  cloud-platform environment rds-drift-checker s3://your-bucket/path/to/merged-rds-errored-namespaces.csv

		Run with a local CSV file:
		  cloud-platform environment rds-drift-checker file://path/to/merged-rds-errored-namespaces.csv
	`),
	Args:   cobra.ExactArgs(1),
	PreRun: upgradeIfNotLatest,
	RunE:   environment.RdsDriftChecker,
}
