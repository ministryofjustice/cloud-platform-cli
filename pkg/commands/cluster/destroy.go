package clustercmd

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func addDestroyClusterCmd(toplevel *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "destroy",
		Short: `destroy a cloud-platform cluster that doesn't have the production tag`,
		Run: func(cmd *cobra.Command, args []string) {
			contextLogger := log.WithFields(log.Fields{"subcommand": "destroy-cluster"})
			if os.Getenv("AWS_PROFILE") == "" || os.Getenv("AWS_ACCESS_KEY_ID") == "" && os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
				contextLogger.Fatal("You must have the following environment variables set: AWS_PROFILE, AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY")
			}

			kubeClient := client.NewKubeClient()

		},
	}

	// check if aws env vars are set

	// authenticate to the cluster

	// check if the cluster exists and is not production

	// destroy the cluster

	toplevel.AddCommand(cmd)
}
