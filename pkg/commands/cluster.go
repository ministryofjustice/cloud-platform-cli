package commands

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type clusterDrainCmd struct {
	Node    string // name of the node to drain
	Force   bool   // force drain and ignore customer uptime requests
	DryRun  bool   // don't actually drain the node
	TimeOut int    // draining a node usually takes around two minutes. If it takes longer than this, it will be cancelled.
}

var c clusterDrainCmd

func addClusterCmd(topLevel *cobra.Command) {
	topLevel.AddCommand(clusterCmd)

	// sub cobra commands
	clusterCmd.AddCommand(clusterDrainNodeCmd)

	// flags
	clusterDrainNodeCmd.Flags().StringVarP(&c.Node, "node", "n", "", "node to drain")
	clusterDrainNodeCmd.Flags().BoolVarP(&c.Force, "force", "f", false, "force drain and ignore customer uptime requests")
	clusterDrainNodeCmd.Flags().BoolVar(&c.DryRun, "dry-run", false, "don't actually drain the node")
	clusterDrainNodeCmd.Flags().IntVarP(&c.TimeOut, "timeout", "t", 360, "draining a node usually takes around two minutes. If it takes longer than this, it will be cancelled.")
}

var clusterCmd = &cobra.Command{
	Use:    "cluster",
	Short:  `Cloud Platform cluster actions`,
	PreRun: upgradeIfNotLatest,
}

var clusterDrainNodeCmd = &cobra.Command{
	Use:   "drain-node",
	Short: `Drain the oldest node from the cluster`,
	Example: heredoc.Doc(`
	$ cloud-platform cluster drain-node
	`),
	PreRun: upgradeIfNotLatest,
	RunE: func(cmd *cobra.Command, args []string) error {

		return nil
	},
}
