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
	environmentCmd.AddCommand(environmentSvcCmd)
	environmentSvcCmd.AddCommand(environmentSvcCreateCmd)
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
	Short: `Creates a serviceaccount`,
	Example: heredoc.Doc(`
	$ cloud-platform environment serviceaccount 
	`),
	// PreRun: upgradeIfNotLatest,
}

var environmentSvcCreateCmd = &cobra.Command{
	Use:   "create",
	Short: `Create serviceaccount to a namespace dir`,
	Example: heredoc.Doc(`
	$ cloud-platform environment serviceaccount create
	`),
	// PreRun: upgradeIfNotLatest,
	RunE: environment.CreateTemplateServiceAccount,
}
