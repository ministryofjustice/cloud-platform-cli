package commands

import (
	"errors"

	"github.com/MakeNowJust/heredoc"
	duplicate "github.com/ministryofjustice/cloud-platform-cli/pkg/duplicate"
	"github.com/spf13/cobra"
)

var (
	DuplicateIngressNamespace string
)

func addDuplicateCmd(topLevel *cobra.Command) {
	topLevel.AddCommand(duplicateCmd)
	duplicateCmd.AddCommand(duplicateIngressCmd)

	duplicateIngressCmd.Flags().StringVarP(&DuplicateIngressNamespace, "namespace", "n", "", "Namespace which you want to perform the duplicate resource")

}

var duplicateCmd = &cobra.Command{
	Use:   "duplicate",
	Short: `Cloud Platform duplicate resource`,
}

var duplicateIngressCmd = &cobra.Command{
	Use:   "ingress",
	Short: `Duplicating ingress for the given ingress resource name and namespace in live cluster`,
	Example: heredoc.Doc(`
	$ cloud-platform duplicate ingress
	`),
	PreRun: upgradeIfNotLatest,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			return nil
		}
		return errors.New("requires existing ingress resource name")

	},
	RunE: func(cmd *cobra.Command, args []string) error {

		return duplicate.DuplicateTestIngress()
	},
}
