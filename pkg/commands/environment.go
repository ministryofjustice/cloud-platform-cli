package commands

import (
	environment "github.com/ministryofjustice/cloud-platform-tools/pkg/environment"

	"github.com/spf13/cobra"
)

func addEnvironmentCmd(topLevel *cobra.Command) {
	topLevel.AddCommand(environmentCmd)
	environmentCmd.AddCommand(environmentRdsCmd)
	environmentRdsCmd.AddCommand(environmentRdsCreateCmd)
}

var environmentCmd = &cobra.Command{
	Use:   "environment",
	Short: `Cloud Platform Environment actions`,
}

var environmentRdsCmd = &cobra.Command{
	Use:   "rds",
	Short: `RDS instances operations, create, list, remove`,
}

var environmentRdsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: `Create terraform files for RDS instance and its related AWS resources`,
	RunE:  environment.CreateTemplateRds,
}
