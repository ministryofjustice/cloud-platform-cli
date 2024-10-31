package commands

import (
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/pipeline"
	util "github.com/ministryofjustice/cloud-platform-cli/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type pipelineOptions struct {
	util.Options
	Name          string
	NodeGroupName string
	MaxNameLength int8
	BranchName    string
}

var cliOpt pipelineOptions

func (opt *pipelineOptions) addPipelineCordonAndDrainClusterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&cliOpt.Name, "cluster-name", "", "cluster to run pipeline cmds against")
	cmd.Flags().StringVar(&cliOpt.NodeGroupName, "node-group", "", "node group name to cordon and drain")
}

func addPipelineCordonAndDrainClusterCmd(toplevel *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "cordon-and-drain",
		Short: `cordon and drain a node group on a cluster`,
		Long: heredoc.Doc(`

			Running this command will cordon and drain an existing node group in a eks cluster in the cloud-platform aws account.
      It will not terminate the nodes nor will it delete the node group.

      The cordon and drain will run remotely in the pipeline, and under the hood it calls "cloud-platform cluster recycle-node --name <node-name> --drain-only --ignore-label"

			You must have the following environment variables set, or passed via arguments:
				- a cluster name

      ** You _must_ have the fly cli installed **
      --> https://concourse-ci.org/fly.html

      ** You must also have wget installed **
      --> brew install wget
`),
		PreRun: upgradeIfNotLatest,
		Run: func(cmd *cobra.Command, args []string) {
			contextLogger := log.WithFields(log.Fields{"subcommand": "cordon-and-drain"})
			cliOpt.MaxNameLength = 12

			if cliOpt.Name == "" {
				contextLogger.Fatal("--cluster-name is required")
			}

			if err := cliOpt.IsNameValid(); err != nil {
				contextLogger.Fatal(err)
			}

			if cliOpt.NodeGroupName == "" {
				contextLogger.Fatal("--node-group is required")
			}

			if strings.Contains(cliOpt.Name, "manager") && strings.Contains(cliOpt.NodeGroupName, "def-ng") {
				contextLogger.Fatal("⚠️ Warning ⚠️ Because this command runs remotely in concourse, this command can’t be used to drain default ng on the manager cluster. It must be run locally while your context is set to the correct cluster. see https://runbooks.cloud-platform.service.justice.gov.uk/node-group-changes.html#process-for-recycling-all-nodes-in-a-cluster for more details")
			}

			pipeline.CordonAndDrainPipelineShellCmds(cliOpt.Name, cliOpt.NodeGroupName)
		},
	}

	cliOpt.addPipelineCordonAndDrainClusterFlags(cmd)
	toplevel.AddCommand(cmd)
}

func (opt *pipelineOptions) addPipelineDeleteClusterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&cliOpt.Name, "cluster-name", "", "cluster to delete")
	cmd.Flags().StringVarP(&cliOpt.BranchName, "branch-name", "b", "main", "branch name to use for pipeline run (default: main)")
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

			Optionally you can pass a branch name to use for the pipeline run, default is "main"

      ** You _must_ have the fly cli installed **
      --> https://concourse-ci.org/fly.html

      ** You must also have wget installed **
      --> brew install wget
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

			pipeline.DeletePipelineShellCmds(cliOpt.Name, cliOpt.BranchName)
		},
	}

	cliOpt.addPipelineDeleteClusterFlags(cmd)
	toplevel.AddCommand(cmd)
}

func addPipelineCmd(topLevel *cobra.Command) {
	topLevel.AddCommand(pipelineCmd)

	addPipelineDeleteClusterCmd(pipelineCmd)
	addPipelineCordonAndDrainClusterCmd(pipelineCmd)
}

var pipelineCmd = &cobra.Command{
	Use:    "pipeline",
	Short:  `Cloud Platform pipeline actions`,
	PreRun: upgradeIfNotLatest,
}
