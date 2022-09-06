package cmd

import (
	"errors"

	commands "github.com/ministryofjustice/cloud-platform-cli/pkg/commands"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var validArgs = []string{"environment", "decode-secret", "duplicate", "kubecfg", "terraform", "version"}

var SkipVersionCheck bool

var RootCmd = &cobra.Command{
	Use: "cloud-platform",
	// validArgs lets us use the short form of the command, e.g. `cloud-platform environment`
	ValidArgs:         validArgs,
	DisableAutoGenTag: true,
	Short:             "Multi-purpose CLI from the Cloud Platform team",
	RunE:              RootCmdRunE,
}

func RootCmdRunE(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	return errors.New("I don't recognise that command. Try `cloud-platform --help`")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	commands.AddCommands(RootCmd)
	RootCmd.AddCommand(docs)

	cobra.CheckErr(RootCmd.Execute())
}

func rootCmdFlags(cmd *cobra.Command) {
	// We need the option to bypass the automatic update process, so that new
	// releases of the cloud-platform tool don't break any pipelines which use
	// the tool. This allows us to add `--skip-version-check` to any command
	// which runs in a pipeline.
	RootCmd.PersistentFlags().BoolVarP(&SkipVersionCheck, "skip-version-check", "", false, "don't check for updates")
	_ = viper.BindPFlag("skip-version-check", RootCmd.PersistentFlags().Lookup("skip-version-check"))
}
func init() {
	rootCmdFlags(RootCmd)
}
