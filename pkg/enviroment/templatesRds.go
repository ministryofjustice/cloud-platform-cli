package enviroment

import (
	"os"
	"text/template"
)

type templateRds struct {
	IsProduction bool
	Environment  string
	BusinessUnit string
	Application  string
	Namespace    string
	TeamName     string
}

// CreateTemplateRds creates the terraform files from environment's template folder
func CreateTemplateRds() error {
	tpl, err := template.ParseGlob("templates/terraform/rds/*")

	rdsValues, err := templateRdsSetValues()
	if err != nil {
		panic(err)
	}

	err = tpl.Execute(os.Stdout, rdsValues)
	if err != nil {
		panic(err)
	}

	return nil
}

func templateRdsSetValues() (*templateRds, error) {
	values := templateRds{}

	err := validatePath()
	if err != nil {
		outsidePath := promptYesNo{label: "WARNING: You are outside the cloud-platform environment. Do you want to continue and render templates on the screen?", defaultValue: 0}
		err = outsidePath.promptyesNo()
		if err != nil {
			return nil, err
		}
	}

	environments, err := GetEnvironmentsFromGH()
	if err != nil {
		panic(err)
	}

	// spew.Dump(environments)

	environmentName, err := promptSelectEnvironments(environments)
	if err != nil {
		return nil, err
	}

	metadata := MetaDataFromGH{environmentName: environmentName}
	err = metadata.GetEnvironmentsMetadataFromGH()
	if err != nil {
		return nil, err
	}

	isProduction := promptYesNo{label: "Is Production?", defaultValue: 0}
	if metadata.isProduction == "false" {
		isProduction.defaultValue = 1
	}

	err = isProduction.promptyesNo()
	if err != nil {
		return nil, err
	}

	application := promptString{label: "Application name?", defaultValue: metadata.application}
	application.promptString()
	if err != nil {
		return nil, err
	}

	businessUnit := promptString{label: "Business Unit?", defaultValue: metadata.businessUnit}
	businessUnit.promptString()
	if err != nil {
		return nil, err
	}

	return &values, nil
}
