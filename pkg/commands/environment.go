package commands

import (
	environment "github.com/ministryofjustice/cloud-platform-cli/pkg/environment"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

func addEnvironmentCmd(topLevel *cobra.Command) {
	topLevel.AddCommand(environmentCmd)
	environmentCmd.AddCommand(environmentEcrCmd)
	environmentCmd.AddCommand(environmentRdsCmd)
	environmentCmd.AddCommand(environmentS3Cmd)
	environmentCmd.AddCommand(environmentSvcCmd)
	environmentCmd.AddCommand(environmentCreateCmd)
	environmentEcrCmd.AddCommand(environmentEcrCreateCmd)
	environmentRdsCmd.AddCommand(environmentRdsCreateCmd)
	environmentS3Cmd.AddCommand(environmentS3CreateCmd)
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

var environmentEcrCmd = &cobra.Command{
	Use:   "ecr",
	Short: `Add an ECR to a namespace`,
	Example: heredoc.Doc(`
	$ cloud-platform environment ecr create
	`),
	PreRun: upgradeIfNotLatest,
}

var environmentEcrCreateCmd = &cobra.Command{
	Use:    "create",
	Short:  `Create "resources/ecr.tf" terraform file for an ECR`,
	PreRun: upgradeIfNotLatest,
	RunE:   environment.CreateTemplateEcr,
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

var environmentS3Cmd = &cobra.Command{
	Use:   "s3",
	Short: `Add a S3 bucket to a namespace`,
	Example: heredoc.Doc(`
	$ cloud-platform environment s3 create
	`),
	PreRun: upgradeIfNotLatest,
}

var environmentS3CreateCmd = &cobra.Command{
	Use:    "create",
	Short:  `Create "resources/s3.tf" terraform file for a S3 bucket`,
	PreRun: upgradeIfNotLatest,
	RunE:   environment.CreateTemplateS3,
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
	$ cloud-platform environment serviceaccount create
	`),
	PreRun: upgradeIfNotLatest,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := environment.CreateTemplateServiceAccount(); err != nil {
			return err
		}

		return nil
	},
}
