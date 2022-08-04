package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/gruntwork-io/go-commons/git"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/recycle"
	"github.com/ministryofjustice/cloud-platform-go-library/client"
	"github.com/ministryofjustice/cloud-platform-go-library/cluster"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

// Global variables used for cluster creation
var (
	createOptions = &cluster.CreateOptions{}
	auth          = &cluster.AuthOpts{}
	date          = time.Now().Format("0201")
	minHour       = time.Now().Format("1504")
)

var recycleOptions recycle.Options

var awsSecret, awsAccessKey, awsProfile string

func addClusterCmd(topLevel *cobra.Command) {
	topLevel.AddCommand(clusterCmd)

	// sub cobra commands
	clusterCmd.AddCommand(clusterRecycleNodeCmd)
	clusterCmd.AddCommand(clusterCreateCmd)

	// recycle node flags
	clusterRecycleNodeCmd.Flags().StringVarP(&recycleOptions.ResourceName, "name", "n", "", "name of the resource to recycle")
	clusterRecycleNodeCmd.Flags().BoolVarP(&recycleOptions.Force, "force", "f", true, "force the pods to drain")
	clusterRecycleNodeCmd.Flags().BoolVarP(&recycleOptions.IgnoreLabel, "ignore-label", "i", false, "whether to ignore the labels on the resource")
	clusterRecycleNodeCmd.Flags().IntVarP(&recycleOptions.TimeOut, "timeout", "t", 360, "amount of time to wait for the drain command to complete")
	clusterRecycleNodeCmd.Flags().BoolVar(&recycleOptions.Oldest, "oldest", false, "whether to recycle the oldest node")
	clusterRecycleNodeCmd.Flags().StringVar(&recycleOptions.KubecfgPath, "kubecfg", filepath.Join(homedir.HomeDir(), ".kube", "config"), "path to kubeconfig file")
	clusterRecycleNodeCmd.Flags().StringVar(&recycleOptions.AwsRegion, "aws-region", "eu-west-2", "aws region to use")
	clusterRecycleNodeCmd.Flags().BoolVar(&recycleOptions.Debug, "debug", false, "enable debug logging")

	// Global cluster flags
	clusterCmd.Flags().StringVar(&awsAccessKey, "aws-access-key", os.Getenv("AWS_ACCESS_KEY_ID"), "aws access key to use")
	clusterCmd.Flags().StringVar(&awsSecret, "aws-secret-key", os.Getenv("AWS_SECRET_ACCESS_KEY"), "aws secret to use")
	clusterCmd.Flags().StringVar(&awsProfile, "aws-profile", os.Getenv("AWS_PROFILE"), "aws profile to use")

	// Add cluster flags
	clusterCreateCmd.Flags().StringVar(&auth.ClientId, "auth0-client-id", os.Getenv("AUTH0_CLIENT_ID"), "[required] auth0 client id to use")
	clusterCreateCmd.Flags().StringVar(&auth.ClientSecret, "auth0-client-secret", os.Getenv("AUTH0_CLIENT_SECRET"), "[required] auth0 client secret to use")
	clusterCreateCmd.Flags().StringVar(&auth.Domain, "auth0-domain", os.Getenv("AUTH0_DOMAIN"), "[required] auth0 domain to use")

	// if a name is not specified, create a random one using the format DD-MM-HH-MM
	clusterCreateCmd.Flags().StringVar(&createOptions.Name, "name", fmt.Sprintf("cp-%s-%s", date, minHour), "[optional] name of the cluster")
	clusterCreateCmd.Flags().StringVar(&createOptions.VpcName, "vpc", createOptions.Name, "[optional] name of the vpc to use")
	clusterCreateCmd.Flags().StringVar(&createOptions.ClusterSuffix, "cluster-suffix", "cloud-platform.service.justice.gov.uk", "[optional] suffix to append to the cluster name")
	clusterCreateCmd.Flags().BoolVar(&createOptions.Debug, "debug", false, "[optional] enable debug logging")
	clusterCreateCmd.Flags().IntVar(&createOptions.NodeCount, "nodes", 3, "[optional] number of nodes to create. [default] 3")
	clusterCreateCmd.Flags().IntVar(&createOptions.TimeOut, "timeout", 600, "[optional] amount of time to wait for the command to complete. [default] 600s")
}

var clusterCmd = &cobra.Command{
	Use:    "cluster",
	Short:  `Cloud Platform cluster actions`,
	PreRun: upgradeIfNotLatest,
}

var clusterCreateCmd = &cobra.Command{
	Use:   "create",
	Short: `Create a new Cloud Platform cluster`,
	Example: heredoc.Doc(`
		$ cloud-platform cluster create --name my-cluster --region eu-west-2 --kubecfg-path ~/.kube/config
	`),
	PreRun: upgradeIfNotLatest,
	RunE: func(cmd *cobra.Command, args []string) error {
		contextLogger := log.WithFields(log.Fields{"subcommand": "create-cluster"})

		// Ensure the executor is in the `cloud-platform-infrastructure` repository
		if err := ensureExecutorInRepository(contextLogger); err != nil {
			return err
		}

		if awsProfile == "" && awsAccessKey == "" && awsSecret == "" {
			contextLogger.Fatal("AWS credentials are required, please set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY or an AWS_PROFILE")
		}

		if createOptions.Auth0.ClientId == "" || createOptions.Auth0.ClientSecret == "" || createOptions.Auth0.Domain == "" {
			contextLogger.Fatal("Auth0 credentials are required, please set AUTH0_CLIENT_ID, AUTH0_CLIENT_SECRET and AUTH0_DOMAIN")
		}

		if len(createOptions.Name) > createOptions.MaxNameLength {
			contextLogger.Fatal("Cluster name is too long, please use a shorter name")
		}

		creds, err := client.NewAwsCreds("eu-west-2")
		if err != nil {
			contextLogger.Fatal(err)
		}

		createOptions.AwsCredentials = *creds

		c := cluster.Cluster{}

		return c.Create(createOptions)
	},
}

var clusterRecycleNodeCmd = &cobra.Command{
	Use:   "recycle-node",
	Short: `recycle a node`,
	Example: heredoc.Doc(`
	$ cloud-platform cluster recycle-node
	`),
	PreRun: upgradeIfNotLatest,
	Run: func(cmd *cobra.Command, args []string) {
// 		contextLogger := log.WithFields(log.Fields{"subcommand": "recycle-node"})
// 		// Check for missing name argument. You must define either a resource
// 		// or specify the --oldest flag.
// 		if recycleOptions.ResourceName == "" && !recycleOptions.Oldest {
// 			contextLogger.Fatal("--name or --oldest is required")
// 		}

// 		if awsProfile == "" && awsAccessKey == "" && awsSecret == "" {
// 			contextLogger.Fatal("AWS credentials are required, please set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY or an AWS_PROFILE")
// 		}

// 		clientset, err := client.GetClientset(recycleOptions.KubecfgPath)
// 		if err != nil {
// 			contextLogger.Fatal(err)
// 		}

// 		recycle := &recycle.Recycler{
// 			Client:  &client.Client{Clientset: clientset},
// 			Options: &recycleOptions,
// 		}

// 		recycle.Cluster, err = cluster.NewCluster(recycle.Client)
// 		if err != nil {
// 			contextLogger.Fatal(err)
// 		}

// 		// Create a snapshot for comparison later.
// 		recycle.Snapshot = recycle.Cluster.NewSnapshot()

// 		recycle.AwsCreds, err = cluster.NewAwsCreds(recycleOptions.AwsRegion)
// 		if err != nil {
// 			contextLogger.Fatal(err)
// 		}

// 		err = recycle.Node()
// 		if err != nil {
// 			// Fail hard so we get an non-zero exit code.
// 			// This is mainly for when this is run in a pipeline.
// 			contextLogger.Fatal(err)
// 		}
// 	},
// }

// ensureExecutorInRepository ensures that the executor is in the `cloud-platform-infrastructure` repository.
func ensureExecutorInRepository(executorPart string) error {
	// Check if the executor is in the `cloud-platform-infrastructure` repository.
	// If not, clone it.
	if _, err := os.Stat(executorPath); os.IsNotExist(err) {
		logger.Info("Executor not found, cloning...")
		if err := git.Clone(executorPath, executorRepo); err != nil {
			return err
		}
	}

	// Check if the executor is in the correct branch.
	// If not, checkout the correct branch.
	if err := git.Checkout(executorPath, executorBranch); err != nil {
		return err
	}

	return nil
}
