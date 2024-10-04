package commands

import (
	"bytes"
	"context"
	"os"
	"strings"

	terraform "github.com/ministryofjustice/cloud-platform-cli/pkg/terraform"
	util "github.com/ministryofjustice/cloud-platform-cli/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func addTerraformCmd(topLevel *cobra.Command) {
	var tf terraform.TerraformCLIConfig
	var diff bool

	rootCmd := &cobra.Command{
		Use:    "terraform",
		Short:  `Terraform actions.`,
		PreRun: upgradeIfNotLatest,
	}

	plan := &cobra.Command{
		Use:    "plan",
		Short:  `Terraform plan.`,
		PreRun: upgradeIfNotLatest,
		Run: func(cmd *cobra.Command, args []string) {
			contextLogger := log.WithFields(log.Fields{"subcommand": "plan"})
			if tf.Workspace == "" {
				contextLogger.Fatal("Workspace is required")
			}

			if tf.IsPipeline && tf.PlanFilename == "" {
				contextLogger.Fatal("When running in the pipeline you must provide a plan filename")
			}

			tfCli, err := terraform.NewTerraformCLI(&tf)
			if err != nil {
				contextLogger.Fatal(err)
			}

			var out bytes.Buffer

			err = tfCli.Init(context.Background(), &out)
			// print terraform init output irrespective of error. out captures both stdout and stderr of terraform
			util.Redacted(os.Stdout, out.String(), tfCli.Redacted)
			if err != nil {
				contextLogger.Fatal("Failed to init terraform: %w", err)
			}

			_, err = tfCli.Plan(context.Background(), &out)
			util.Redacted(os.Stdout, out.String(), tfCli.Redacted)
			if err != nil {
				contextLogger.Fatal("Failed to plan terraform: %w", err)
			}
		},
	}

	apply := &cobra.Command{
		Use:    "apply",
		Short:  `Terraform apply.`,
		PreRun: upgradeIfNotLatest,
		Run: func(cmd *cobra.Command, args []string) {
			contextLogger := log.WithFields(log.Fields{"subcommand": "apply"})
			if tf.Workspace == "" {
				contextLogger.Fatal("Workspace is required")
			}

			if tf.IsPipeline && tf.PlanFilename == "" {
				contextLogger.Fatal("When running in the pipeline you must provide a plan filename")
			}

			tfCli, err := terraform.NewTerraformCLI(&tf)
			if err != nil {
				contextLogger.Fatal(err)
			}

			var out bytes.Buffer

			err = tfCli.Init(context.Background(), &out)
			// print terraform init output irrespective of error. out captures both stdout and stderr of addTerraformCmd
			util.Redacted(os.Stdout, out.String(), tfCli.Redacted)
			if err != nil {
				contextLogger.Fatal("Failed to init terraform: %w", err)
			}

			err = tfCli.Apply(context.Background(), &out)
			util.Redacted(os.Stdout, out.String(), tfCli.Redacted)
			if err != nil {
				contextLogger.Fatal("Failed to apply terraform: %w", err)
			}
		},
	}

	checkDivergence := &cobra.Command{
		Use:    "check-divergence",
		Short:  `Terraform check-divergence check if there are drifts in the state.`,
		PreRun: upgradeIfNotLatest,
		Run: func(cmd *cobra.Command, args []string) {
			contextLogger := log.WithFields(log.Fields{"subcommand": "check-divergence"})
			if tf.Workspace == "" {
				contextLogger.Fatal("Workspace is required")
			}

			contextLogger.Info("Executing terraform plan, if there is a drift this program execution will fail")

			tfCli, err := terraform.NewTerraformCLI(&tf)
			if err != nil {
				contextLogger.Fatal(err)
			}

			var out bytes.Buffer
			err = tfCli.Init(context.Background(), &out)
			// print terraform init output irrespective of error. out captures both stdout and stderr of terraform
			util.Redacted(os.Stdout, out.String(), tfCli.Redacted)
			if err != nil {
				contextLogger.Fatal("Failed to init terraform: %w", err)
			}

			// diff - false if there is are changes in the plan
			diff, err = tfCli.Plan(context.Background(), &out)
			util.Redacted(os.Stdout, out.String(), tfCli.Redacted)
			if err != nil {
				contextLogger.Fatal("Failed to plan terraform: %w", err)
			}

			if diff {
				contextLogger.Fatal("There is a drift when executing terraform plan")
			}
		},
	}

	addCommonFlags(checkDivergence, &tf)
	addCommonFlags(plan, &tf)
	addCommonFlags(apply, &tf)
	rootCmd.AddCommand(checkDivergence)
	rootCmd.AddCommand(plan)
	rootCmd.AddCommand(apply)
	topLevel.AddCommand(rootCmd)
}

func addCommonFlags(cmd *cobra.Command, tf *terraform.TerraformCLIConfig) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	cmd.PersistentFlags().StringVarP(&awsAccessKey, "aws-access-key-id", "", "", "[required] Access key id of service account to be used by terraform")
	cmd.PersistentFlags().StringVarP(&awsSecret, "aws-secret-access-key", "", "", "[required] Secret access key of service account to be used by terraform")
	cmd.PersistentFlags().StringVarP(&awsRegion, "aws-region", "", "", "[required] aws region to use")
	// Terraform options
	cmd.PersistentFlags().StringVar(&tf.WorkingDir, "workdir", ".", "[optional] the terraform working directory to perform terraform operation [default] .")
	cmd.PersistentFlags().StringVarP(&tf.Workspace, "workspace", "w", "", "[required] workspace where terraform is going to be executed")
	cmd.PersistentFlags().StringVar(&tf.Version, "terraform-version", "1.2.5", "[optional] the terraform version to use.")
	cmd.PersistentFlags().BoolVar(&tf.Redacted, "redact", true, "Redact the terraform output before printing")
	cmd.PersistentFlags().BoolVar(&tf.IsPipeline, "is-pipeline", false, "[required] if the terraform is being executed from the pipeline")
	cmd.PersistentFlags().StringVar(&tf.PlanFilename, "plan-filename", "", "[optional] the plan filename to be output from the terraform plan or used for the terraform apply eg. 'plan-$PR_NUM.out' [default] ''")

	_ = cmd.MarkPersistentFlagRequired("aws-access-key-id")
	_ = cmd.MarkPersistentFlagRequired("aws-secret-access-key")
	_ = cmd.MarkPersistentFlagRequired("aws-region")
	_ = cmd.MarkPersistentFlagRequired("workspace")

	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			_ = cmd.PersistentFlags().Set(f.Name, viper.GetString(f.Name))
		}
	})
}
