package commands

import (
	"os"

	"github.com/MakeNowJust/heredoc"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/kubectl/pkg/cmd/clusterinfo"
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

			// Print kubeconfig location

			// Print cluster-info

			kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
			matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)

			f := cmdutil.NewFactory(matchVersionKubeConfigFlags)
			o := &clusterinfo.ClusterInfoOptions{
				IOStreams: genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
			}
			var err error
			o.Client, err = f.ToRESTConfig()
			if err != nil {
				contextLogger.Fatal("Failed to get clusterInfo: %w", err)
			}

			o.Builder = resource.NewBuilder(f)
			cmdutil.CheckErr(o.Run())

			// print cluster name

			// print list of namespaces

		},
	}

	//whereamiCmd.Flags().StringVar(&kubePath, "kubecfg", filepath.Join(homedir.HomeDir(), ".kube", "config"), "path to kubeconfig file")
	topLevel.AddCommand(whereamiCmd)
}
