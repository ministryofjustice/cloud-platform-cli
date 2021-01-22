package commands

import (
	"github.com/MakeNowJust/heredoc"
	decodeSecret "github.com/ministryofjustice/cloud-platform-cli/pkg/decodeSecret"
	"github.com/spf13/cobra"
)

func addDecodeSecret(topLevel *cobra.Command) {
	opts := &decodeSecret.DecodeSecretOptions{}

	cmd := &cobra.Command{
		Use:   "decode-secret",
		Short: `Decode a kubernetes secret`,
		Example: heredoc.Doc(`
$ cloud-platform decode-secret -n mynamespace -s mysecret
	`),
		PreRun: upgradeIfNotLatest,
		RunE: func(cmd *cobra.Command, args []string) error {
			return decodeSecret.DecodeSecret(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Secret, "secret", "s", "", "Secret name")
	cmd.MarkFlagRequired("secret")

	cmd.Flags().StringVarP(&opts.Namespace, "namespace", "n", "", "Namespace name")
	cmd.MarkFlagRequired("namespace")

	cmd.Flags().BoolVarP(&opts.ExportAwsCreds, "export-aws-credentials", "e", false, "Export AWS credentials as shell variables")

	topLevel.AddCommand(cmd)
}
