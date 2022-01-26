package commands

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/MakeNowJust/heredoc"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/client"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/cluster"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/recycle"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

var opt recycle.Options

func addClusterCmd(topLevel *cobra.Command) {
	topLevel.AddCommand(clusterCmd)

	// sub cobra commands
	clusterCmd.AddCommand(clusterRecycleNodeCmd)

	// recycle node flags
	clusterRecycleNodeCmd.Flags().StringVarP(&opt.ResourceName, "name", "n", "", "name of the resource to recycle")
	clusterRecycleNodeCmd.Flags().IntVarP(&opt.TimeOut, "timeout", "t", 360, "amount of time to wait for the drain command to complete")
	clusterRecycleNodeCmd.Flags().BoolVar(&opt.Oldest, "oldest", false, "whether to recycle the oldest node")
	clusterRecycleNodeCmd.Flags().StringVar(&opt.KubecfgPath, "kubecfg", filepath.Join(homedir.HomeDir(), ".kube", "config"), "path to kubeconfig file")
	clusterRecycleNodeCmd.Flags().StringVar(&opt.AwsProfile, "aws-profile", "default", "aws profile to use")
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
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check for missing name argument. You must define either a resource
		// or specify the --oldest flag.
		if opt.ResourceName == "" && !opt.Oldest {
			return errors.New("--name or --oldest is required")
		}

		clientset, err := client.GetClientset(opt.KubecfgPath)
		if err != nil {
			return err
		}

		recycle := &recycle.Recycler{
			Client:  &client.Client{Clientset: clientset},
			Options: &opt,
		}

		recycle.Cluster, err = cluster.NewCluster(recycle.Client)
		if err != nil {
			return fmt.Errorf("failed to get cluster: %s", err)
		}

		// Create a snapshot for comparison later.
		recycle.Snapshot = recycle.Cluster.NewSnapshot()

		err = recycle.Node()
		if err != nil {
			err := recycle.RemoveLabel("node-cordon")
			if err != nil {
				return fmt.Errorf("failed to remove node-cordon label: %s", err)
			}
			err = recycle.RemoveLabel("node-drain")
			if err != nil {
				return fmt.Errorf("failed to remove node-drain label: %s", err)
			}

			return err
		}

		return nil
	},
}
