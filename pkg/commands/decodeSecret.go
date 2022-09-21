package commands

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	decodeSecret "github.com/ministryofjustice/cloud-platform-cli/pkg/decodeSecret"
	"github.com/ministryofjustice/cloud-platform-go-library/client"
	"github.com/spf13/cobra"
)

var (
	secret         string
	namespace      string
	exportAwsCreds bool
	raw            bool
)

func addDecodeSecret(topLevel *cobra.Command) {
	topLevel.AddCommand(decodeSecretCmd)

	decodeSecretCmd.Flags().StringVarP(&secret, "secret", "s", "", "Secret name")
	_ = decodeSecretCmd.MarkFlagRequired("secret")

	decodeSecretCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Namespace name")
	_ = decodeSecretCmd.MarkFlagRequired("namespace")

	decodeSecretCmd.Flags().BoolVarP(&exportAwsCreds, "export-aws-credentials", "e", false, "Export AWS credentials as shell variables")

	decodeSecretCmd.Flags().BoolVarP(&raw, "raw", "r", false, "Output the raw secret, rather than prettyprinting")
}

var decodeSecretCmd = &cobra.Command{
	Use:   "decode-secret",
	Short: `Decode a kubernetes secret`,
	Example: heredoc.Doc(`
$ cloud-platform decode-secret -n mynamespace -s mysecret
	`),
	PreRun: upgradeIfNotLatest,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create clientset
		clientset, err := client.NewKubeClientWithValues(kubeconfig, "")
		if err != nil {
			return fmt.Errorf("error creating clientset: %v", err)
		}

		// Create options
		opts, err := decodeSecret.NewOptions(clientset.Clientset, secret, namespace, exportAwsCreds, raw)
		if err != nil {
			return fmt.Errorf("error creating options: %v", err)
		}

		s := decodeSecret.Secret{}

		return s.DecodeSecret(opts)
	},
}
