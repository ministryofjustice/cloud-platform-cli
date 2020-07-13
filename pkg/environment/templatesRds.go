package environment

import (
	"fmt"
	"strconv"
	"text/template"

	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

type templateRds struct {
	IsProduction          bool
	EnvironmentName       string
	BusinessUnit          string
	Application           string
	Namespace             string
	InfrastructureSupport string
	RdsModuleName         string
	TeamName              string
}

// CreateTemplateRds creates the terraform files from environment's template folder
func CreateTemplateRds(cmd *cobra.Command, args []string) error {

	RdsTemplate, err := downloadTemplate("https://raw.githubusercontent.com/ministryofjustice/cloud-platform-terraform-rds-instance/template-patch/template/rds.tmpl")
	if err != nil {
		return (err)
	}

	rdsValues, err := templateRdsSetValues()
	if err != nil {
		return (err)
	}

	tpl := template.Must(template.New("rds").Parse(RdsTemplate))

	outputPath := fmt.Sprintf("namespaces/live-1.cloud-platform.service.justice.gov.uk/%s/resources/rds.tf", rdsValues.Namespace)
	f, _ := outputFileWriter(outputPath)
	err = tpl.Execute(f, rdsValues)
	if err != nil {
		return (err)
	}

	fmt.Printf("RDS File generated in namespaces/live-1.cloud-platform.service.justice.gov.uk/%s/resources/rds.tf\n", rdsValues.Namespace)
	color.Info.Tips("This template is using default values provided by your namespace information. Please review before raising PR")

	return nil
}

func templateRdsSetValues() (*templateRds, error) {
	values := templateRds{}
	metadata := metadataFromNamespace{}

	_, err := metadata.checkPath()
	if err != nil {
		return nil, err
	}

	err = metadata.getNamespaceFromPath()
	if err != nil {
		return nil, err
	}

	err = metadata.checkNamespaceExist()
	if err != nil {
		return nil, err
	}

	err = metadata.getNamespaceMetadata()
	if err != nil {
		return nil, err
	}

	values.Application = metadata.application
	values.Namespace = metadata.namespace
	values.BusinessUnit = metadata.businessUnit
	values.EnvironmentName = metadata.environmentName
	values.IsProduction, _ = strconv.ParseBool(metadata.isProduction)
	values.RdsModuleName = "rds"
	values.InfrastructureSupport = metadata.ownerEmail
	values.TeamName = "teamName"

	return &values, nil
}
