package commands

import (
	"github.com/MakeNowJust/heredoc"
	bump "github.com/ministryofjustice/cloud-platform-cli/pkg/bump"
	"github.com/spf13/cobra"
)

func bumpCmd(topLevel *cobra.Command) {
	opts := &bump.BumpOptions{}

	cmd := &cobra.Command{
		Use:   "bump",
		Short: `bump all users release version`,
		Example: heredoc.Doc(`
cloud-platform bump --module serviceaccount --version 1.1.1

Would bump all users release in the environments repository to the specified version.
	`),
		PreRun: upgradeIfNotLatest,
		RunE: func(cmd *cobra.Command, args []string) error {
			return bump.ModuleVersion(*opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Module, "module", "m", "", "module name")
	cmd.MarkFlagRequired("module")

	cmd.Flags().StringVarP(&opts.Version, "version", "v", "", "version to change to")
	cmd.MarkFlagRequired("version")

	cmd.Flags().StringVarP(&opts.Namespace, "namespace", "n", "", "namespace name")

	topLevel.AddCommand(cmd)
}
