package clustercmd

import (
	"fmt"
	"os"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/cluster"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type Destroy struct {
	Name       string
	Force      bool
	DestroyVpc bool
}

func addDestroyClusterCmd(toplevel *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "destroy",
		Short: `destroy a cloud-platform cluster that doesn't have the production tag`,
		Run: func(cmd *cobra.Command, args []string) {

			contextLogger := log.WithFields(log.Fields{"subcommand": "destroy-cluster"})
			if os.Getenv("AWS_PROFILE") == "" || os.Getenv("AWS_ACCESS_KEY_ID") == "" && os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
				contextLogger.Fatal("You must have the following environment variables set: AWS_PROFILE, AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY")
			}

			var clusterName string
			var force, destroyVpc bool
			cmd.Flags().StringVar(&clusterName, "cluster-name", "", "[required] name of the cluster to destroy")
			cmd.Flags().BoolVar(&force, "force", false, "force the destruction of the cluster")
			cmd.Flags().BoolVar(&destroyVpc, "destroy-vpc", false, "destroy the vpc associated with the cluster")

			if err := cmd.MarkFlagRequired("cluster-name"); err != nil {
				contextLogger.Fatal(err)
			}

			if err := cmd.Execute(); err != nil {
				contextLogger.Fatal(err)
			}

			destroy := newDestory(clusterName, force, destroyVpc)
			if err := destroy.Validate(); err != nil {
				contextLogger.Fatal(err)
			}

			if err := destroy.Run(); err != nil {
				contextLogger.Fatal(err)
			}

			kubeClient, err := cluster.NewKubeClient("")
			if err != nil {
				contextLogger.Fatal(err)
			}
			fmt.Println(kubeClient)

		},
	}

	// check if aws env vars are set

	// authenticate to the cluster

	// check if the cluster exists and is not production

	// destroy the cluster

	toplevel.AddCommand(cmd)
}

func newDestory(name string, force bool, destroyVpc bool) *Destroy {
	// does the cluster exist?
	// is the cluster production?

	return &Destroy{
		Name:       name,
		Force:      force,
		DestroyVpc: destroyVpc,
	}
}
func (d *Destroy) Run() error {
	return nil
}

func (d *Destroy) Validate() error {
	return nil
}
