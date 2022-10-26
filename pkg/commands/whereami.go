package commands

import (
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/tools/clientcmd"
	clusterinfo "k8s.io/kubectl/pkg/cmd/clusterinfo"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func addWhereamiCmd(topLevel *cobra.Command) {
	alias := []string{"where", "am", "i"}
	whereamiCmd := &cobra.Command{
		Use:     "whereami",
		Aliases: alias,
		Short:   `Information on the cluster based on the kubeconfig`,
		Example: heredoc.Doc(`
$ cloud-platform whereami
	`),
		PreRun: upgradeIfNotLatest,
		Run: func(cmd *cobra.Command, args []string) {
			contextLogger := log.WithFields(log.Fields{"command": "whereami"})

			// Lookup for KUBECONFIG environment variable
			kubePath, ok := os.LookupEnv("KUBECONFIG")
			if !ok {
				fmt.Printf("KUBECONFIG not set\n")
			} else {
				fmt.Printf("Your KUBECONFIG environment variable is set to %s\n", kubePath)
			}
			// Print current context
			err := printCurrentContext(kubePath)
			if err != nil {
				contextLogger.Fatal("Failed to get current context: %w", err)
			}
			// Print cluster-info
			err = printClusterInfo(kubePath)

			if err != nil {
				contextLogger.Fatal("Failed to get clusterInfo: %w", err)
			}

			// print list of namespaces

		},
	}

	whereamiCmd.Flags().StringVar(&kubePath, "kubecfg", os.Getenv("KUBECONFIG"), "path of kubeconfig set as KUBECONFIG environment variable")
	topLevel.AddCommand(whereamiCmd)
}

func printCurrentContext(kubepath string) error {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubepath},
		&clientcmd.ConfigOverrides{
			CurrentContext: "",
		}).RawConfig()

	if config.CurrentContext == "" {
		err = fmt.Errorf("current-context is not set")
		return err
	}

	fmt.Printf("Your current context is %s\n", config.CurrentContext)
	return nil

}

// func printGetContext(kubepath string) error {
// 	// Print kubeconfig location

// 	contextObj := &config.GetContextsOptions{
// 		ConfigAccess: clientcmd.NewDefaultPathOptions(),
// 		IOStreams:    genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
// 	}

// 	pathOptions := clientcmd.NewDefaultPathOptions()
// 	pathOptions.GlobalFile = kubePath
// 	pathOptions.EnvVar = ""
// 	contextObj.configAccess = clientcmd.NewDefaultPathOptions()

// 	cmdutil.CheckErr(contextObj.RunGetContexts())
// 	return nil
// }

func printClusterInfo(kubePath string) error {
	kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)

	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	clusterInfoObj := &clusterinfo.ClusterInfoOptions{
		IOStreams: genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
	}

	var err error
	clusterInfoObj.Client, err = f.ToRESTConfig()
	if err != nil {
		return err
	}
	clusterInfoObj.Builder = resource.NewBuilder(f)
	cmdutil.CheckErr(clusterInfoObj.Run())
	return nil
}
