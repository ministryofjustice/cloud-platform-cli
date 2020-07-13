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
	Short: `RDS instances operations, create, list, remove`,
	Example: heredoc.Doc(`
	$ cloud-platform environment rds create
	`),
}

var environmentRdsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: `Create terraform files for RDS instance and its related AWS resources`,
	RunE:  environment.CreateTemplateRds,
}
