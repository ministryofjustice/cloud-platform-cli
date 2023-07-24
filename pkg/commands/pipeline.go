package commands

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/pipeline"
	util "github.com/ministryofjustice/cloud-platform-cli/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type pipelineOptions struct {
	util.Options
	Name          string
	MaxNameLength int8
}

var cliOpt pipelineOptions

func (opt *pipelineOptions) addPipelineDeleteClusterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&cliOpt.Name, "cluster-name", "", "cluster to delete")
}

func addPipelineDeleteClusterCmd(toplevel *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "delete-cluster",
		Short: `delete a cloud-platform cluster via the pipeline`,
		Long: heredoc.Doc(`

			Running this command will delete an existing eks cluster in the cloud-platform aws account.
			It will delete the components, then the cluster, the VPC, terraform workspace and cleanup the cloudwatch log group.

      The delete will run remotely in the pipeline

			You must have the following environment variables set, or passed via arguments:
				- a cluster name

      ** You _must_ have the fly cli installed **
      --> https://concourse-ci.org/fly.html
`),
		PreRun: upgradeIfNotLatest,
		Run: func(cmd *cobra.Command, args []string) {
			contextLogger := log.WithFields(log.Fields{"subcommand": "delete-cluster"})
			cliOpt.MaxNameLength = 12

			if cliOpt.Name == "" {
				contextLogger.Fatal("--cluster-name is required")
			}

			if err := cliOpt.IsNameValid(); err != nil {
				contextLogger.Fatal(err)
			}

			pipeline.DeletePipelineShellCmds(cliOpt.Name)
		},
	}

	cliOpt.addPipelineDeleteClusterFlags(cmd)
	toplevel.AddCommand(cmd)
}

func addPipelineCmd(topLevel *cobra.Command) {
	topLevel.AddCommand(pipelineCmd)

	addPipelineDeleteClusterCmd(pipelineCmd)
}

var pipelineCmd = &cobra.Command{
	Use:    "pipeline",
	Short:  `Cloud Platform pipeline actions`,
	PreRun: upgradeIfNotLatest,
}
