package commands

import (
	"strings"

	terraform "github.com/ministryofjustice/cloud-platform-cli/pkg/terraform"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func addTerraformCmd(topLevel *cobra.Command) {

	var options terraform.Commander

	rootCmd := &cobra.Command{
		Use:    "terraform",
		Short:  `Terraform actions.`,
		PreRun: upgradeIfNotLatest,
	}

	checkDivergence := &cobra.Command{
		Use:    "check-divergence",
		Short:  `Terraform check-divergence check if there are drifts in the state.`,
		PreRun: upgradeIfNotLatest,
		Run: func(cmd *cobra.Command, args []string) {
			contextLogger := log.WithFields(log.Fields{"subcommand": "check-divergence"})

			contextLogger.Info("Executing terraform plan, if there is a drift this program execution will fail")
			err := options.CheckDivergence()

			if err != nil {
				contextLogger.Fatal("Error executing plan, either an error or a divergence")
			}
		},
	}

	apply := &cobra.Command{
		Use:    "apply",
		Short:  `Execute terraform apply.`,
		PreRun: upgradeIfNotLatest,
		Run: func(cmd *cobra.Command, args []string) {
			contextLogger := log.WithFields(log.Fields{"subcommand": "apply"})

			contextLogger.Info("Executing terraform apply")
			err := options.Apply()

			if err != nil {
				contextLogger.Fatal("Error executing terraform apply - check the outputs")
			}
		},
	}

	plan := &cobra.Command{
		Use:    "plan",
		Short:  `Execute terraform plan.`,
		PreRun: upgradeIfNotLatest,
		Run: func(cmd *cobra.Command, args []string) {
			contextLogger := log.WithFields(log.Fields{"subcommand": "plan"})

			if options.BulkTfPlanPaths == "" {
				contextLogger.Info("Executing terraform plan")
				err := options.Plan()

				if err != nil {
					contextLogger.Fatal("Error executing terraform plan - check the outputs")
				}
			} else {
				err := options.BulkPlan()

				if err != nil {
					contextLogger.Fatal(err)
				}

			}

		},
	}

	addCommonFlags(checkDivergence, &options)
	addCommonFlags(apply, &options)
	addCommonFlags(plan, &options)
	rootCmd.AddCommand(plan)
	rootCmd.AddCommand(apply)
	rootCmd.AddCommand(checkDivergence)
	topLevel.AddCommand(rootCmd)
}

func addCommonFlags(cmd *cobra.Command, o *terraform.Commander) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	cmd.PersistentFlags().StringVarP(&o.AccessKeyID, "aws-access-key-id", "", "", "Access key id of service account to be used by terraform")
	cmd.PersistentFlags().StringVarP(&o.SecretAccessKey, "aws-secret-access-key", "", "", "Secret access key of service account to be used by terraform")
	cmd.PersistentFlags().StringVarP(&o.Workspace, "workspace", "w", "default", "Default workspace where terraform is going to be executed")
	cmd.PersistentFlags().BoolVarP(&o.DisplayTfOutput, "display-tf-output", "d", true, "Display or not terraform plan output")
	cmd.PersistentFlags().StringVarP(&o.VarFile, "var-file", "v", "", "tfvar to be used by terraform")
	cmd.PersistentFlags().StringVar(&o.BulkTfPlanPaths, "dirs-file", "", "Required for bulk-plans, file path which holds directories where terraform plan is going to be executed")
	cmd.PersistentFlags().StringVar(&o.Context, "context", "", "kops or eks?")

	cmd.MarkPersistentFlagRequired("aws-access-key-id")
	cmd.MarkPersistentFlagRequired("aws-secret-access-key")

	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			cmd.PersistentFlags().Set(f.Name, viper.GetString(f.Name))
		}
	})
}
