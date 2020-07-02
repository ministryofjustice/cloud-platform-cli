package commands

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

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
	fmt.Println(options.namespace)
	getEnvironmentsFromGithub()
}

// GetEnvironmentsFromGithub returns the environments names from Cloud Platform Environments
// repository (in Github)
func getEnvironmentsFromGithub() {
	response, err := http.Get("https://raw.githubusercontent.com/ministryofjustice/cloud-platform-terraform-rds-instance/main/example/rds.tf")
	if err != nil {
		fmt.Println(err)
	}
	body, _ := ioutil.ReadAll(response.Body)
	text := string(body)
	fmt.Println(text)

	response, err := http.Get("https://raw.githubusercontent.com/ministryofjustice/cloud-platform-terraform-rds-instance/main/example/rds.tf")
	if err != nil {
		return nil, err
	}
	data, _ := ioutil.ReadAll(response.Body)
	content := string(data)
	fmt.Println(content)

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
