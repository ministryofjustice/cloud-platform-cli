package commands

import (
	"strings"

	enviroment "github.com/ministryofjustice/cloud-platform-tools/pkg/enviroment"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type PoornimaOptions struct {
	message string
}

func addPoornima(topLevel *cobra.Command) {

	options := PoornimaOptions{}

	poornimaCmd := &cobra.Command{
		Use:   "poornima",
		Short: `This is a test for poornima.`,
		Run: func(cmd *cobra.Command, args []string) {

			enviroment.CreateTemplateRds()
		},
	}

	addCommonFlagsPoornima(poornimaCmd, &options)

	topLevel.AddCommand(poornimaCmd)
}

func addCommonFlagsPoornima(cmd *cobra.Command, o *PoornimaOptions) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	cmd.PersistentFlags().StringVarP(&o.message, "message", "", "", "This is a message which is going to be printed")

	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			cmd.PersistentFlags().Set(f.Name, viper.GetString(f.Name))
		}
	})
}
