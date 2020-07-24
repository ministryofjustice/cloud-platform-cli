package commands

import (
	environment "github.com/ministryofjustice/cloud-platform-cli/pkg/environment"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

func addEnvironmentCmd(topLevel *cobra.Command) {
	topLevel.AddCommand(environmentCmd)
	environmentCmd.AddCommand(environmentRdsCmd)
	environmentCmd.AddCommand(environmentCreateCmd)
	environmentRdsCmd.AddCommand(environmentRdsCreateCmd)
}

var environmentCmd = &cobra.Command{
	Use:   "environment",
	Short: `Cloud Platform Environment actions`,
}

var environmentCreateCmd = &cobra.Command{
	Use:   "create",
	Short: `Create an environment`,
	Example: heredoc.Doc(`
	$ cloud-platform environment create
	`),
	RunE: environment.CreateTemplateNamespace,
}

var environmentRdsCmd = &cobra.Command{
	Use:   "rds",
	Short: `Add an RDS instance to a namespace`,
	Example: heredoc.Doc(`
	$ cloud-platform environment rds create
	`),
}

var environmentRdsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: `Create "resources/rds.tf" terraform file for an RDS instance`,
	RunE:  environment.CreateTemplateRds,
}
