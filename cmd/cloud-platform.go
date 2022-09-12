package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	commands "github.com/ministryofjustice/cloud-platform-cli/pkg/commands"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var validArgs = []string{"environment", "decode-secret", "duplicate", "kubecfg", "terraform", "version"}

var SkipVersionCheck bool

var rootCmd = &cobra.Command{
	Use: "cloud-platform",
	// validArgs lets us use the short form of the command, e.g. `cloud-platform environment`
	ValidArgs:         validArgs,
	DisableAutoGenTag: true,
	Short:             "Multi-purpose CLI from the Cloud Platform team",
	RunE:              RootCmdRunE,
}

func RootCmdRunE(cmd *cobra.Command, args []string) error {
	_, err := cmd.Flags().GetBool("skip-version-check")
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return cmd.Help()
	}

	return errors.New("I don't recognise that command. Try `cloud-platform --help`")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	commands.AddCommands(rootCmd)
	rootCmd.AddCommand(docs)

	cobra.CheckErr(rootCmd.Execute())
}

func RootCmdFlags(cmd *cobra.Command) {
	// We need the option to bypass the automatic update process, so that new
	// releases of the cloud-platform tool don't break any pipelines which use
	// the tool. This allows us to add `--skip-version-check` to any command
	// which runs in a pipeline.
	rootCmd.PersistentFlags().BoolVarP(&SkipVersionCheck, "skip-version-check", "", false, "don't check for updates")
	_ = viper.BindPFlag("skip-version-check", rootCmd.PersistentFlags().Lookup("skip-version-check"))
}
func init() {
	RootCmdFlags(rootCmd)
}

func ExecuteCommand(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)

	args = append([]string{"--skip-version-check"}, args...)

	err := cmd.Execute()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), err
}
