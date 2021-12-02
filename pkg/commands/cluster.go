package commands

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

func addClusterCmd(topLevel *cobra.Command) {
	topLevel.AddCommand(clusterCmd)
	clusterCmd.AddCommand(clusterDrainNodeCmd)
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
}
