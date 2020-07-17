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

const RdsTemplateFile = "https://raw.githubusercontent.com/ministryofjustice/cloud-platform-terraform-rds-instance/main/template/rds.tmpl"
const RdsTfFile = "resources/rds.tf"

// CreateTemplateRds creates the terraform files from environment's template folder
func CreateTemplateRds(cmd *cobra.Command, args []string) error {

	RdsTemplate, err := downloadTemplate(RdsTemplateFile)
	if err != nil {
		return (err)
	}

	rdsValues, err := templateRdsSetValues()
	if err != nil {
		return (err)
	}

	tpl := template.Must(template.New("rds").Parse(RdsTemplate))

	f, _ := outputFileWriter(RdsTfFile)
	err = tpl.Execute(f, rdsValues)
	if err != nil {
		return (err)
	}

	fmt.Printf("RDS File generated in %s\n", RdsTfFile)
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

	namespace := Namespace{}
	namespace.ReadYaml()

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
