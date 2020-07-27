package commands

import (
	"fmt"
	"strings"

	terraform "github.com/ministryofjustice/cloud-platform-cli/pkg/terraform"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type terraformOptions struct {
	accessKeyID     string
	secretAccessKey string
	workspace       string
	varFile         string
	displayTfOutput bool
}

func addTerraformCmd(topLevel *cobra.Command) {

	var options terraformOptions

	rootCmd := &cobra.Command{
		Use:    "terraform",
		Short:  `Terraform actions.`,
		PreRun: upgradeIfNotLatest,
	}

	checkDivergence := &cobra.Command{
		Use:   "check-divergence",
		Short: `Terraform check-divergence check if there are drifts in the state.`,
		PreRun: upgradeIfNotLatest,
		Run: func(cmd *cobra.Command, args []string) {
			terraformCheckDivergence(&options)
		},
	}

	addCommonFlags(checkDivergence, &options)
	rootCmd.AddCommand(checkDivergence)
	topLevel.AddCommand(rootCmd)
}

// terraformCheckDivergence tell us if there is a divergence in our terraform plan.
// This basically translates in: "is there any changes pending to apply?"
func terraformCheckDivergence(o *terraformOptions) error {

	contextLogger := log.WithFields(log.Fields{"subcommand": "check-divergence"})

	terraform := terraform.Commander{}

	if o.displayTfOutput {
		terraform.DisplayOutput = true
	}

	var TfVarFile []string

	// Check if user provided a terraform var-file
	if o.varFile != "" {
		TfVarFile = append([]string{fmt.Sprintf("-var-file=%s", o.varFile)})
	}

	contextLogger.Info("Executing terraform plan, if there is a drift this program execution will fail")

	err := terraform.CheckDivergence(o.workspace, TfVarFile...)

	if err != nil {
		contextLogger.Fatal("Error executing plan, either an error or a divergence")
	}

	return nil
}

func addCommonFlags(cmd *cobra.Command, o *terraformOptions) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	cmd.PersistentFlags().StringVarP(&o.accessKeyID, "aws-access-key-id", "", "", "Access key id of service account to be used by terraform")
	cmd.PersistentFlags().StringVarP(&o.secretAccessKey, "aws-secret-access-key", "", "", "Secret access key of service account to be used by terraform")
	cmd.PersistentFlags().StringVarP(&o.workspace, "workspace", "w", "default", "Default workspace where terraform is going to be executed")
	cmd.PersistentFlags().BoolVarP(&o.displayTfOutput, "display-tf-output", "d", true, "Display or not terraform plan output")
	cmd.PersistentFlags().StringVarP(&o.varFile, "var-file", "v", "", "tfvar to be used by terraform")

	cmd.MarkPersistentFlagRequired("aws-access-key-id")
	cmd.MarkPersistentFlagRequired("aws-secret-access-key")

	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			cmd.PersistentFlags().Set(f.Name, viper.GetString(f.Name))
		}
	})
}
