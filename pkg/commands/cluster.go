package commands

import (
	"path/filepath"

	"github.com/MakeNowJust/heredoc"
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
	clusterRecycleNodeCmd.Flags().StringVarP(&opt.NodeName, "node", "n", "", "node to recycle")
	clusterRecycleNodeCmd.Flags().IntVarP(&opt.TimeOut, "timeout", "t", 360, "draining a node usually takes around two minutes. If it takes longer than this, it will be cancelled.")
	clusterRecycleNodeCmd.Flags().BoolVar(&opt.Oldest, "oldest", false, "whether to recycle the oldest node")
	clusterRecycleNodeCmd.Flags().StringVar(&opt.Kubecfg, "kubecfg", filepath.Join(homedir.HomeDir(), ".kube", "config"), "path to kubeconfig file")
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
	Short: `choose a node to recycle`,
	Example: heredoc.Doc(`
	$ cloud-platform cluster recycle-node
	`),
	PreRun: upgradeIfNotLatest,
	RunE: func(cmd *cobra.Command, args []string) error {
		var r = recycle.New()

		err := r.Node(&opt)
		if err != nil {
			return err
		}

		return nil
	},
}
