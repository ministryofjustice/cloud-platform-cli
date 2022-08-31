package main

import (
	"flag"
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

	commands.AddCommands(cmds)

	// We need the option to bypass the automatic update process, so that new
	// releases of the cloud-platform tool don't break any pipelines which use
	// the tool. This allows us to add `--skip-version-check` to any command
	// which runs in a pipeline.
	var SkipVersionCheck bool
	cmds.PersistentFlags().BoolVarP(&SkipVersionCheck, "skip-version-check", "", false, "don't check for updates")
	_ = viper.BindPFlag("skip-version-check", cmds.PersistentFlags().Lookup("skip-version-check"))

	// Document generation is usually left to an automated pipeline, but can be
	// triggered manually using the --generate-docs flag.
	var GenerateDocs bool
	// Using the stdlib flag package as we need to parse the flag before cmd.Execute() is called.
	flag.BoolVar(&GenerateDocs, "generate-docs", false, "generate CLI documentation")
	flag.Parse()

	if GenerateDocs {
		fmt.Println("Generating docs...")
		if _, err := os.Stat("doc"); os.IsNotExist(err) {
			log.Fatalln("doc directory does not exist, assuming we're not in the cli repository")
		}
		if err := doc.GenMarkdownTree(cmds, "./doc"); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	if err := cmds.Execute(); err != nil {
		fmt.Printf("Error during command execution: %v", err)
		os.Exit(0)
	}
}
