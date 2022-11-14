package commands

import (
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:    "check",
	Short:  `Check the validity of a configuration`,
	PreRun: upgradeIfNotLatest,
}

func addCheckCmd(toplevel *cobra.Command) {
	toplevel.AddCommand(checkCmd)

	checkCmd.AddCommand(ingressCheckCmd())
}

func ingressCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ingress",
		Short: "Check the validity of an ingress",
	}
	return cmd
}
