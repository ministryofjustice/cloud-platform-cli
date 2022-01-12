package commands

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/cluster"
)

var c cluster.RecycleNodeOpt

func addClusterCmd(topLevel *cobra.Command) {
	topLevel.AddCommand(clusterCmd)

	// sub cobra commands
	clusterCmd.AddCommand(clusterRecycleNodeCmd)

	// flags
	clusterRecycleNodeCmd.Flags().StringVarP(&c.Node.Name, "node", "n", "", "node to recycle")
	clusterRecycleNodeCmd.Flags().BoolVarP(&c.Force, "force", "f", false, "force drain and ignore customer uptime requests")
	clusterRecycleNodeCmd.Flags().BoolVar(&c.DryRun, "dry-run", false, "don't actually recycle the node")
	clusterRecycleNodeCmd.Flags().IntVarP(&c.TimeOut, "timeout", "t", 360, "draining a node usually takes around two minutes. If it takes longer than this, it will be cancelled.")
	clusterRecycleNodeCmd.Flags().BoolVar(&c.Oldest, "oldest", false, "whether to recycle the oldest node")
	clusterRecycleNodeCmd.Flags().StringVar(&c.KubeConfigPath, "kubecfg", "", "path to kubeconfig file")
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
		err := cluster.RecycleNode(&c)
		if err != nil {
			return err
		}

		return nil
	},
}
