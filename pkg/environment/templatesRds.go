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

	re := RepoEnvironment{}
	err := re.MustBeInCloudPlatformEnvironments()
	if err != nil {
		return nil, err
	}

	ns := Namespace{}
	err = ns.ReadYaml()
	if err != nil {
		return nil, err
	}

	values.Application = ns.application
	values.Namespace = ns.name
	values.BusinessUnit = ns.businessUnit
	values.EnvironmentName = ns.environmentName
	values.IsProduction, _ = strconv.ParseBool(ns.isProduction)
	values.RdsModuleName = "rds"
	values.InfrastructureSupport = ns.ownerEmail
	values.TeamName = "teamName"

	return &values, nil
}
