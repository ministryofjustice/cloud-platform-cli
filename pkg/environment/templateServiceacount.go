package environment

import (
	"fmt"

	"github.com/gookit/color"
)

const (
	svcAccTemplateFile = "https://raw.githubusercontent.com/ministryofjustice/cloud-platform-terraform-serviceaccount/main/template/serviceaccount.tmpl"
	svcAccTfFile       = "resources/serviceaccount.tf"
)

// CreateTemplateServiceAccount sets and creates a template file containing all
// the necessary values to create a serviceaccount resource in Kubernetes. It
// will only execute in a directory with a namespace resource i.e. 00-namespace.yaml.
func CreateTemplateServiceAccount() error {
	re := RepoEnvironment{}
	err := re.mustBeInANamespaceFolder()
	if err != nil {
		return err
	}

	err = createSvcAccTfFile()
	if err != nil {
		return err
	}

	fmt.Println(svcAccTfFile, "created")
	fmt.Printf("Serviceaccount File generated in %s\n", svcAccTfFile)
	color.Info.Tips("Please review before raising PR")

	return nil
}

//------------------------------------------------------------------------------

func createSvcAccTfFile() error {
	// The serviceaccount "template" is actually an example file that we can just save
	// as is into the user's resources/ directory as `serviceaccount.tf`
	return CopyUrlToFile(svcAccTemplateFile, svcAccTfFile)
}
