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

	RdsTemplate, err := downloadTemplate("https://raw.githubusercontent.com/ministryofjustice/cloud-platform-terraform-rds-instance/main/template/rds.tmpl")
	if err != nil {
		return (err)
	}

	rdsValues, err := templateRdsSetValues()
	if err != nil {
		return (err)
	}

	tpl := template.Must(template.New("rds").Parse(RdsTemplate))

	outputPath := fmt.Sprintf("%s/%s/resources/rds.tf", namespaceBaseFolder, rdsValues.Namespace)
	f, _ := outputFileWriter(outputPath)
	err = tpl.Execute(f, rdsValues)
	if err != nil {
		return (err)
	}

	fmt.Printf("RDS File generated in %s/%s/resources/rds.tf\n", namespaceBaseFolder, rdsValues.Namespace)
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

	namespace := Namespace{}
	namespace.ReadYamlFile("00-namespace.yaml")

	values.Application = namespace.application
	values.Namespace = namespace.name
	values.BusinessUnit = namespace.businessUnit
	values.EnvironmentName = namespace.environmentName
	values.IsProduction, _ = strconv.ParseBool(namespace.isProduction)
	values.RdsModuleName = "rds"
	values.InfrastructureSupport = namespace.ownerEmail
	values.TeamName = "teamName"

	return &values, nil
}
