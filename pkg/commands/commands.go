package commands

import (
	"github.com/spf13/cobra"
)

// AddCommands is a function to group all commands
func AddCommands(topLevel *cobra.Command) {
	addTerraformCmd(topLevel)
	addKubecfgCmd(topLevel)
	addVersion(topLevel)
	addEnvironmentCmd(topLevel)
	addPrototypeCmd(topLevel)
	addDecodeSecret(topLevel)
	addDuplicateCmd(topLevel)
	addClusterCmd(topLevel)
	addPipelineCmd(topLevel)
}
