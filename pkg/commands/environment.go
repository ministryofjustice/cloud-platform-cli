package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type rdsCreateOptions struct {
	namespace string
}

func addEnvironmentCmd(topLevel *cobra.Command) {

	var options rdsCreateOptions

	rootCmd := &cobra.Command{
		Use:   "environment",
		Short: `Cloud Platform Environment actions.`,
	}

	resourceRDSCmd := &cobra.Command{
		Use:   "rds",
		Short: `RDS instance operations, create, list, remove`,
	}

	rdsCreateCmd := &cobra.Command{
		Use:   "create",
		Short: `Create terraform files for RDS instance and its related AWS resources`,
		Run: func(cmd *cobra.Command, args []string) {
			environmentRDSCreate(&options)
		},
	}

	addRDSCreateCommonFlags(rdsCreateCmd, &options)

	resourceRDSCmd.AddCommand(rdsCreateCmd)
	rootCmd.AddCommand(resourceRDSCmd)
	topLevel.AddCommand(rootCmd)

}

func environmentRDSCreate(options *rdsCreateOptions) {
	fmt.Println(options.namespace)
}

func addRDSCreateCommonFlags(cmd *cobra.Command, o *rdsCreateOptions) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	cmd.PersistentFlags().StringVarP(&o.namespace, "namespace", "", "", "Namespace for which RDS has to be created")
	cmd.MarkPersistentFlagRequired("namespace")

	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			cmd.PersistentFlags().Set(f.Name, viper.GetString(f.Name))
		}
	})
}
