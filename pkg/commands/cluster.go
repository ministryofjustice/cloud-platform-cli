package commands

import (
	"os"
	"path/filepath"

	"github.com/MakeNowJust/heredoc"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/client"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/cluster"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/recycle"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

var opt recycle.Options

var awsSecret, awsAccessKey, awsProfile, awsRegion string

func addClusterCmd(topLevel *cobra.Command) {
	topLevel.AddCommand(clusterCmd)

	// sub cobra commands
	clusterCmd.AddCommand(clusterRecycleNodeCmd)

	// recycle node flags
	clusterRecycleNodeCmd.Flags().StringVarP(&opt.ResourceName, "name", "n", "", "name of the resource to recycle")
	clusterRecycleNodeCmd.Flags().BoolVarP(&opt.Force, "force", "f", true, "force the pods to drain")
	clusterRecycleNodeCmd.Flags().BoolVarP(&opt.IgnoreLabel, "ignore-label", "i", false, "whether to ignore the labels on the resource")
	clusterRecycleNodeCmd.Flags().IntVarP(&opt.TimeOut, "timeout", "t", 360, "amount of time to wait for the drain command to complete")
	clusterRecycleNodeCmd.Flags().BoolVar(&opt.Oldest, "oldest", false, "whether to recycle the oldest node")
	clusterRecycleNodeCmd.Flags().StringVar(&opt.KubecfgPath, "kubecfg", filepath.Join(homedir.HomeDir(), ".kube", "config"), "path to kubeconfig file")
	clusterRecycleNodeCmd.Flags().StringVar(&awsAccessKey, "aws-access-key", os.Getenv("AWS_ACCESS_KEY_ID"), "aws access key to use")
	clusterRecycleNodeCmd.Flags().StringVar(&awsSecret, "aws-secret-key", os.Getenv("AWS_SECRET_ACCESS_KEY"), "aws secret to use")
	clusterRecycleNodeCmd.Flags().StringVar(&awsProfile, "aws-profile", os.Getenv("AWS_PROFILE"), "aws profile to use")
	clusterRecycleNodeCmd.Flags().StringVar(&opt.AwsRegion, "aws-region", "eu-west-2", "aws region to use")
	clusterRecycleNodeCmd.Flags().BoolVar(&opt.Debug, "debug", false, "enable debug logging")
}

var clusterCmd = &cobra.Command{
	Use:    "cluster",
	Short:  `Cloud Platform cluster actions`,
	PreRun: upgradeIfNotLatest,
}

var clusterRecycleNodeCmd = &cobra.Command{
	Use:   "recycle-node",
	Short: `recycle a node`,
	Example: heredoc.Doc(`
	$ cloud-platform cluster recycle-node
	`),
	PreRun: upgradeIfNotLatest,
	Run: func(cmd *cobra.Command, args []string) {
		contextLogger := log.WithFields(log.Fields{"subcommand": "recycle-node"})
		// Check for missing name argument. You must define either a resource
		// or specify the --oldest flag.
		if opt.ResourceName == "" && !opt.Oldest {
			contextLogger.Fatal("--name or --oldest is required")
		}

		if awsProfile == "" && awsAccessKey == "" && awsSecret == "" {
			contextLogger.Fatal("AWS credentials are required, please set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY or an AWS_PROFILE")
		}

		clientset, err := client.GetClientset(opt.KubecfgPath)
		if err != nil {
			contextLogger.Fatal(err)
		}

		recycle := &recycle.Recycler{
			Client:  &client.Client{Clientset: clientset},
			Options: &opt,
		}

		recycle.Cluster, err = cluster.NewCluster(recycle.Client)
		if err != nil {
			contextLogger.Fatal(err)
		}

		// Create a snapshot for comparison later.
		recycle.Snapshot = recycle.Cluster.NewSnapshot()

		recycle.AwsCreds, err = cluster.NewAwsCreds(opt.AwsRegion)
		if err != nil {
			contextLogger.Fatal(err)
		}

		err = recycle.Node()
		if err != nil {
			// Fail hard so we get an non-zero exit code.
			// This is mainly for when this is run in a pipeline.
			contextLogger.Fatal(err)
		}
	},
}
