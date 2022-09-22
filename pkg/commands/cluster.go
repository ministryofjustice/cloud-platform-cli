package commands

import (
	"errors"
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
	Run:    recycleNode,
}

func addClusterCmd(topLevel *cobra.Command) {
	topLevel.AddCommand(clusterCmd)

	// sub cobra commands
	clusterCmd.AddCommand(clusterRecycleNodeCmd)
}

func recycleNode(cmd *cobra.Command, args []string) {
	// Opt will be used to collect all options from the executor.
	var opt recycle.Options
	// awsSecret and awsAccessKey are used to collect the aws credentials from the executor.
	// These are required to create the aws client and will be used to terminate instances.
	var awsSecret, awsAccessKey, awsProfile string

	// Standard logger that seems to be used through the cli.
	contextLogger := log.WithFields(log.Fields{"subcommand": "recycle-node"})

	// Gather arguments from the user.
	if err := applyRecycleFlags(cmd, &opt, awsSecret, awsAccessKey, awsProfile); err != nil {
		contextLogger.Fatal(err)
	}

	// A kubernetes clientset is required to interact with the cluster.
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

	// A snapshot is used to compare expected and actual state.
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
}

func applyRecycleFlags(cmd *cobra.Command, opt *recycle.Options, awsSecret, awsAccessKey, awsProfile string) error {
	cmd.Flags().StringVarP(&opt.ResourceName, "name", "n", "", "name of the resource to recycle")
	cmd.Flags().BoolVarP(&opt.Force, "force", "f", true, "force the pods to drain")
	cmd.Flags().BoolVarP(&opt.IgnoreLabel, "ignore-label", "i", false, "whether to ignore the labels on the resource")
	cmd.Flags().IntVarP(&opt.TimeOut, "timeout", "t", 360, "amount of time to wait for the drain command to complete")
	cmd.Flags().BoolVar(&opt.Oldest, "oldest", false, "whether to recycle the oldest node")
	cmd.Flags().StringVar(&opt.KubecfgPath, "kubecfg", filepath.Join(homedir.HomeDir(), ".kube", "config"), "path to kubeconfig file")
	cmd.Flags().StringVar(&awsAccessKey, "aws-access-key", os.Getenv("AWS_ACCESS_KEY_ID"), "aws access key to use")
	cmd.Flags().StringVar(&awsSecret, "aws-secret-key", os.Getenv("AWS_SECRET_ACCESS_KEY"), "aws secret to use")
	cmd.Flags().StringVar(&awsProfile, "aws-profile", os.Getenv("AWS_PROFILE"), "aws profile to use")
	cmd.Flags().StringVar(&opt.AwsRegion, "aws-region", "eu-west-2", "aws region to use")
	cmd.Flags().BoolVar(&opt.Debug, "debug", false, "enable debug logging")
	// validate
	if err := validateRecycleFlags(opt, awsSecret, awsAccessKey, awsProfile); err != nil {
		return err
	}

	return nil
}

func validateRecycleFlags(opt *recycle.Options, awsSecret, awsAccessKey, awsProfile string) error {
	if opt.ResourceName == "" && !opt.Oldest {
		return errors.New("--name or --oldest is required")
	}

	if awsProfile == "" && awsAccessKey == "" && awsSecret == "" {
		return errors.New("aws credentials are required, please set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY or an AWS_PROFILE")
	}

	return nil
}
