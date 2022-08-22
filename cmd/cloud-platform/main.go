package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/viper"

	commands "github.com/ministryofjustice/cloud-platform-cli/pkg/commands"
)

func main() {
	cmds := &cobra.Command{
		Use:   "cloud-platform",
		Short: "Multi-purpose CLI from the Cloud Platform team",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	// We need the option to bypass the automatic update process, so that new
	// releases of the cloud-platform tool don't break any pipelines which use
	// the tool. This allows us to add `--skip-version-check` to any command
	// which runs in a pipeline.
	var SkipVersionCheck bool
	cmds.PersistentFlags().BoolVarP(&SkipVersionCheck, "skip-version-check", "", false, "don't check for updates")
	_ = viper.BindPFlag("skip-version-check", cmds.PersistentFlags().Lookup("skip-version-check"))

	commands.AddCommands(cmds)

	if err := cmds.Execute(); err != nil {
		fmt.Printf("Error during command execution: %v", err)
		os.Exit(0)
	}

	if err := doc.GenMarkdownTree(cmds, "./doc"); err != nil {
		log.Fatal(err)
	}
}
