package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	commands "github.com/ministryofjustice/cloud-platform-tools/pkg/commands"
)

func main() {
	cmds := &cobra.Command{
		Use:   "cp-tools",
		Short: "Internal multi-purpose CLI for the Cloud Platform team",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	commands.AddCommands(cmds)

	if err := cmds.Execute(); err != nil {
		fmt.Printf("Error during command execution: %v", err)
		os.Exit(0)
	}
}
