package commands

import (
	cluster "github.com/ministryofjustice/cloud-platform-cli/pkg/commands/cluster"
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
	cluster.AddClusterCmd(topLevel)
}
