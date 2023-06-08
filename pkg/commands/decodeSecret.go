package commands

import (
	"github.com/MakeNowJust/heredoc"
	decodeSecret "github.com/ministryofjustice/cloud-platform-cli/pkg/decodeSecret"
	"github.com/spf13/cobra"
)

func addDecodeSecret(topLevel *cobra.Command) {
	opts := &decodeSecret.DecodeSecretOptions{}

	alias := []string{"decode", "secret"}
	cmd := &cobra.Command{
		Use:     "decode-secret",
		Aliases: alias,
		Short:   `Decode a kubernetes secret`,
		Example: heredoc.Doc(`
$ cloud-platform decode-secret -n mynamespace -s mysecret
$ cloud-platform decode-secret -s mysecret  [if you are setting namespace via kubectl context]
	`),
		PreRun: upgradeIfNotLatest,
		RunE: func(cmd *cobra.Command, args []string) error {
			return decodeSecret.DecodeSecret(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Secret, "secret", "s", "", "Secret name")
	_ = cmd.MarkFlagRequired("secret")

	cmd.Flags().StringVarP(&opts.Namespace, "namespace", "n", "", "Namespace name")

	cmd.Flags().BoolVarP(&opts.ExportAwsCreds, "export-aws-credentials", "e", false, "Export AWS credentials as shell variables")

	cmd.Flags().BoolVarP(&opts.Raw, "raw", "r", false, "Output the raw secret, rather than prettyprinting")

	topLevel.AddCommand(cmd)
}
