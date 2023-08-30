package environment

import (
	"fmt"

	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

const (
	s3TemplateFile = "https://raw.githubusercontent.com/ministryofjustice/cloud-platform-terraform-s3-bucket/main/examples/s3.tf"
	s3TfFile       = "resources/s3.tf"
)

// CreateTemplateRds creates the terraform files from environment's template folder
func CreateTemplateS3(cmd *cobra.Command, args []string) error {
	re := RepoEnvironment{}
	err := re.mustBeInANamespaceFolder()
	if err != nil {
		return err
	}

	err = createS3TfFile()
	if err != nil {
		return err
	}

	fmt.Printf("S3 File generated in %s\n", s3TfFile)
	color.Info.Tips("This template is using default values provided by your namespace information. Please review before raising PR")

	return nil
}

//------------------------------------------------------------------------------

func createS3TfFile() error {
	// The s3 "template" is actually an example file that we can just save
	// "as is" into the user's resources/ directory as `s3.tf`
	return CopyUrlToFile(s3TemplateFile, s3TfFile)
}
