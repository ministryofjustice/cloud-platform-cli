package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var docs = &cobra.Command{
	Use:               "generate-docs",
	Short:             "Generate markdown docs for the CLI",
	Hidden:            true,
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Generating docs...")
		if _, err := os.Stat("doc"); os.IsNotExist(err) {
			log.Fatalln("doc directory does not exist, assuming we're not in the cli repository")
		}
		if err := doc.GenMarkdownTree(rootCmd, "./doc"); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	},
}
