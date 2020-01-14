package commands

import (
	"github.com/spf13/cobra"
)

// AddCommands ... bla bla bla
func AddCommands(topLevel *cobra.Command) {
	addVersion(topLevel)
	addTerraformCmd(topLevel)
}
