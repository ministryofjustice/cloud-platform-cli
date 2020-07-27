package commands

import (
	"fmt"
	"log"
	"os/user"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	kubecfg "github.com/ministryofjustice/cloud-platform-cli/pkg/kubecfg"
)

func addKubecfgCmd(topLevel *cobra.Command) {
	topLevel.AddCommand(kubecfgCmd)
	kubecfgCmd.AddCommand(kubecfgShowGithubTeamsCmd)

	kubecfgCommonFlags(kubecfgShowGithubTeamsCmd)

}

var kubecfgCmd = &cobra.Command{
	Use:   "kubecfg",
	Short: `Cloud Platform kubeconfig related commands`,
}

var kubecfgShowGithubTeamsCmd = &cobra.Command{
	Use:   "show-github-teams",
	Short: `Checking the GitHub groups in your ID token`,
	Example: heredoc.Doc(`
	$ cloud-platform kubecfg show-github-teams
	`),
	RunE: func(cmd *cobra.Command, args []string) error {
		kubeconfig, err := cmd.Flags().GetString("kubeconfig")
		if err != nil {
			return err
		}

		if err = kubecfg.ShowGithubTeams(kubeconfig); err != nil {
			return err
		}

		return nil
	},
}

func kubecfgCommonFlags(cmd *cobra.Command) {
	usr, err := user.Current()
	if err != nil {
		log.Fatal("Error identifying user's home directory")
	}

	cmd.Flags().StringP("kubeconfig", "f", fmt.Sprintf("%s/.kube/config", usr.HomeDir),
		"Supply a custom kubeconfig file")
}
