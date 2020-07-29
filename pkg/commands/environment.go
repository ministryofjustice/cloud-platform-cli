package commands

import (
	"strings"

	environment "github.com/ministryofjustice/cloud-platform-cli/pkg/environment"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

func addEnvironmentCmd(topLevel *cobra.Command) {
	topLevel.AddCommand(environmentCmd)
	environmentCmd.AddCommand(environmentRdsCmd)
	environmentCmd.AddCommand(environmentSvcCmd)
	environmentCmd.AddCommand(environmentCreateCmd)
	environmentRdsCmd.AddCommand(environmentRdsCreateCmd)
	environmentSvcCmd.AddCommand(environmentSvcCreateCmd)
	environmentSvcCreateCmd.Flags().StringP("name", "n", "cloud-platform-user", "The name of the ServiceAccount resource")

}

var environmentCmd = &cobra.Command{
	Use:    "environment",
	Short:  `Cloud Platform Environment actions`,
	PreRun: upgradeIfNotLatest,
}

var environmentCreateCmd = &cobra.Command{
	Use:   "create",
	Short: `Create an environment`,
	Example: heredoc.Doc(`
	$ cloud-platform environment create
	`),
	PreRun: upgradeIfNotLatest,
	RunE:   environment.CreateTemplateNamespace,
}

var environmentRdsCmd = &cobra.Command{
	Use:   "rds",
	Short: `Add an RDS instance to a namespace`,
	Example: heredoc.Doc(`
	$ cloud-platform environment rds create
	`),
	PreRun: upgradeIfNotLatest,
}

var environmentRdsCreateCmd = &cobra.Command{
	Use:    "create",
	Short:  `Create "resources/rds.tf" terraform file for an RDS instance`,
	PreRun: upgradeIfNotLatest,
	RunE:   environment.CreateTemplateRds,
}

var environmentSvcCmd = &cobra.Command{
	Use:   "serviceaccount",
	Short: `Add a serviceaccount to a namespace`,
	Example: heredoc.Doc(`
	$ cloud-platform environment serviceaccount 
	`),
	PreRun: upgradeIfNotLatest,
}

var environmentSvcCreateCmd = &cobra.Command{
	Use:   "create",
	Short: `Creates a serviceaccount in your chosen namespace`,
	Example: heredoc.Doc(`
	$ cloud-platform environment serviceaccount create [-n <name>]
	`),
	PreRun: upgradeIfNotLatest,
	RunE: func(cmd *cobra.Command, args []string) error {
		nameFlag, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}

		if err = environment.CreateTemplateServiceAccount(strings.ToLower(nameFlag)); err != nil {
			return err
		}

		return nil
	},
}
