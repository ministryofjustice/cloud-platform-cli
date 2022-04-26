package commands

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/terminalcfg"
	"github.com/spf13/cobra"
)

func addTerminalCmd(topLevel *cobra.Command) {
	topLevel.AddCommand(terminalCmd)
	terminalCmd.AddCommand(terminalCfgTestEnvCmd)
	terminalCmd.AddCommand(terminalCfgLiveEnvCmd)
	terminalCmd.AddCommand(terminalCfgManagerEnvCmd)
}

var terminalCmd = &cobra.Command{
	Use:    "terminal",
	Short:  `Terminal setup for kubecxt and terraform workspace`,
	PreRun: upgradeIfNotLatest,
}

var terminalCfgTestEnvCmd = &cobra.Command{
	Use:   "test",
	Short: `sets up a terminal for the chosen test environment`,
	Example: heredoc.Doc(`
	$ cloud-platform terminal test
	`),
	PreRun: upgradeIfNotLatest,
	Run: func(cmd *cobra.Command, args []string) {
		terminalcfg.TestEnv()
	},
}

var terminalCfgLiveEnvCmd = &cobra.Command{
	Use:   "live",
	Short: `sets up a terminal for the live environment`,
	Example: heredoc.Doc(`
	$ cloud-platform terminal live
	`),
	PreRun: upgradeIfNotLatest,
	Run: func(cmd *cobra.Command, args []string) {
		terminalcfg.LiveManagerEnv("live")
	},
}

var terminalCfgManagerEnvCmd = &cobra.Command{
	Use:   "manager",
	Short: `sets up a terminal for the manager environment`,
	Example: heredoc.Doc(`
	$ cloud-platform terminal manager
	`),
	PreRun: upgradeIfNotLatest,
	Run: func(cmd *cobra.Command, args []string) {
		terminalcfg.LiveManagerEnv("manager")
	},
}
