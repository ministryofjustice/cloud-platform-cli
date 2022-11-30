package clustercmd

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/client"
	cloudPlatform "github.com/ministryofjustice/cloud-platform-cli/pkg/cluster"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/recycle"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var opt recycle.Options

var clusterRecycleNodeCmd = &cobra.Command{
	Use:   "recycle-node",
	Short: `recycle a node`,
	Example: heredoc.Doc(`
	$ cloud-platform cluster recycle-node
	`),
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
			Client:  &client.KubeClient{Clientset: clientset},
			Options: &opt,
		}

		recycle.Cluster, err = cloudPlatform.NewCluster(recycle.Client)
		if err != nil {
			contextLogger.Fatal(err)
		}

		// Create a snapshot for comparison later.
		recycle.Snapshot = recycle.Cluster.NewSnapshot()

		recycle.AwsCreds, err = cloudPlatform.NewAwsCreds(opt.AwsRegion)
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
