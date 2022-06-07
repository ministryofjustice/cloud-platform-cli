package commands

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/prototype"
	"github.com/spf13/cobra"
)

var SkipDockerFiles bool

func addPrototypeCmd(topLevel *cobra.Command) {
	topLevel.AddCommand(prototypeCmd)
	prototypeCmd.AddCommand(prototypeDeployCmd)
	prototypeDeployCmd.AddCommand(prototypeDeployCreateCmd)
	prototypeCmd.Flags().BoolVarP(&SkipDockerFiles, "skip-docker-files", "s", false, "Whether to skip the files required to build the docker image i.e Dockerfile, .dockerignore, start.sh")
}

var prototypeCmd = &cobra.Command{
	Use:    "prototype",
	Short:  `Cloud Platform Prototype actions`,
	PreRun: upgradeIfNotLatest,
}

var prototypeDeployCmd = &cobra.Command{
	Use:    "deploy",
	Short:  `Cloud Platform Environment actions`,
	PreRun: upgradeIfNotLatest,
}

var prototypeDeployCreateCmd = &cobra.Command{
	Use:   "create",
	Short: `Create the deployment files and github actions required to deploy in Cloud Platform`,
	Long: `
	Create the deployment files and github actions required to deploy the Prototype kit from a github repository in Cloud Platform.

The files will be generated based on where the current local branch of the prototype github repository is pointed to:

  https://[namespace name]-[branch-name].apps.live.cloud-platform.service.justice.gov.uk

A continuous deployment workflow will be created in the github repository such
that any changes to the branch are deployed to the cloud platform.
	`,
	Example: heredoc.Doc(`
	$ cloud-platform prototype deploy
	`),
	PreRun: upgradeIfNotLatest,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := prototype.CreateDeploymentPrototype(SkipDockerFiles); err != nil {
			return err
		}

		return nil
	},
}
