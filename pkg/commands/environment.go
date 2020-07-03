package commands

import (
	"strings"

	environment "github.com/ministryofjustice/cloud-platform-tools/pkg/environment"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type rdsCreateOptions struct {
	namespace string
}

func addEnvironmentCmd(topLevel *cobra.Command) {

	rootCmd := &cobra.Command{
		Use:   "environment",
		Short: `Cloud Platform Environment actions.`,
	}

	resourceRDSCmd := &cobra.Command{
		Use:   "rds",
		Short: `RDS instance operations, create, list, remove`,
	}

	environmentRDSCmd := addEnvironmentRDSCmd()

	resourceRDSCmd.AddCommand(environmentRDSCmd)
	rootCmd.AddCommand(resourceRDSCmd)
	topLevel.AddCommand(rootCmd)

}

func addEnvironmentRDSCmd() *cobra.Command {

	var options rdsCreateOptions

	rdsCreateCmd := &cobra.Command{
		Use:   "create",
		Short: `Create terraform files for RDS instance and its related AWS resources`,
		Run: func(cmd *cobra.Command, args []string) {
			environmentRDSCreate(&options)
		},
	}

	addRDSCreateCommonFlags(rdsCreateCmd, &options)

	return rdsCreateCmd

}
func environmentRDSCreate(options *rdsCreateOptions) {
	// fmt.Println(options.namespace)
	environment.CreateTemplateRds()
}

func addRDSCreateCommonFlags(cmd *cobra.Command, o *rdsCreateOptions) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	cmd.PersistentFlags().StringVarP(&o.namespace, "namespace", "", "", "Namespace for which RDS has to be created")

	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			cmd.PersistentFlags().Set(f.Name, viper.GetString(f.Name))
		}
	})
}
