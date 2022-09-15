package commands

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/MakeNowJust/heredoc"
	duplicate "github.com/ministryofjustice/cloud-platform-cli/pkg/duplicate"
	"github.com/ministryofjustice/cloud-platform-go-library/client"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

var ns string

var kubeconfig *string

func addDuplicateCmd(topLevel *cobra.Command) {
	topLevel.AddCommand(duplicateCmd)
	duplicateCmd.AddCommand(duplicateIngressCmd)

	if home := homedir.HomeDir(); home != "" {
		duplicateIngressCmd.Flags().StringVarP(kubeconfig, "kubeconfig", "n", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		duplicateIngressCmd.Flags().StringVarP(kubeconfig, "kubeconfig", "n", "", "absolute path to the kubeconfig file")
	}

	duplicateIngressCmd.Flags().StringVarP(&ns, "namespace", "n", "", "Namespace which you want to perform the duplicate resource")
}

var duplicateCmd = &cobra.Command{
	Use:   "duplicate",
	Short: `Cloud Platform duplicate resource`,
}

var duplicateIngressCmd = &cobra.Command{
	Use:   "ingress <ingress name>",
	Short: `Duplicate ingress for the given ingress resource name and namespace`,
	Long: `Gets the ingress resource for the given name and namespace from the cluster,
copies it, change the ingress name and external-dns annotations for the weighted policy and
apply the duplicated ingress to the same namespace.

This command will access the cluster to get the ingress resource and to create the duplicate ingress.
To access the cluster, it assumes that the user has either set the env variable KUBECONFIG to the filepath of kubeconfig or stored the file in the default location ~/.kube/config
	`,
	Example: heredoc.Doc(`
	$ cloud-platform duplicate ingress myingressname -n mynamespace

	`),
	PreRun: upgradeIfNotLatest,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			return nil
		}
		return errors.New("requires existing ingress resource name")
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create clientset
		clientset, err := client.NewKubeClientWithValues(*kubeconfig, "")
		if err != nil {
			return fmt.Errorf("error creating clientset: %v", err)
		}

		// Create duplicate ingress resource
		ingress, err := duplicate.NewIngress(&clientset.Clientset, ns, args[0])
		if err != nil {
			return fmt.Errorf("error creating duplicate ingress: %v", err)
		}

		return ingress.CreateDuplicate()
	},
}
