package environment

import (
	"fmt"

	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

const ecrTemplateFile = "https://raw.githubusercontent.com/ministryofjustice/cloud-platform-terraform-ecr-credentials/main/examples/ecr.tf"
const ecrTfFile = "resources/ecr.tf"

func CreateTemplateEcr(cmd *cobra.Command, args []string) error {
	re := RepoEnvironment{}
	err := re.mustBeInANamespaceFolder()
	if err != nil {
		return err
	}

	err = createEcrTfFile()
	if err != nil {
		return err
	}

	fmt.Printf("ECR File generated in %s\n", ecrTfFile)
	color.Info.Tips("This template is using default values provided by your namespace information. Please review before raising PR")

	return nil
}

//------------------------------------------------------------------------------

func createEcrTfFile() error {
	// The ecr "template" is actually an example file that we can just save
	// "as is" into the user's resources/ directory as `ecr.tf`
	return copyUrlToFile(ecrTemplateFile, ecrTfFile)
}
