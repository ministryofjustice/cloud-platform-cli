package environment

import (
	"fmt"
	"os"

	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

const rdsTemplateFile = "https://raw.githubusercontent.com/ministryofjustice/cloud-platform-terraform-rds-instance/main/example/rds-postgresql.tf"
const rdsTfFile = "resources/rds.tf"

// CreateTemplateRds creates the terraform files from environment's template folder
func CreateTemplateRds(cmd *cobra.Command, args []string) error {
	re := RepoEnvironment{}
	err := re.mustBeInANamespaceFolder()
	if err != nil {
		return err
	}

	err = createRdsTfFile()
	if err != nil {
		return err
	}

	fmt.Printf("RDS File generated in %s\n", rdsTfFile)
	color.Info.Tips("This template is using default values provided by your namespace information. Please review before raising PR")

	return nil
}

//------------------------------------------------------------------------------

func createRdsTfFile() error {
	// The rds "template" is actually an example file that we can just save
	// "as is" into the user's resources/ directory as `rds.tf`
	return copyUrlToFile(rdsTemplateFile,rdsTfFile)
}
