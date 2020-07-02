package enviroment

import (
	"os"
	"text/template"

	"github.com/davecgh/go-spew/spew"
)

type templateRds struct {
	IsProduction bool
	Environment  string
	BusinessUnit string
	Application  string
	Namespace    string
	TeamName     string
}

// func pullTheFileFromTheFollowingURL() {
// URL:
// 	github / blawldsfasdfasf / asdfasfda
// }

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

	spew.Dump(application)

	// businessUnit, err := promptString("Business Unit?")
	// if err != nil {
	// 	return nil, err
	// }
	// values.BusinessUnit = businessUnit

	// teamName, err := promptString("Application name?")
	// if err != nil {
	// 	return nil, err
	// }
	// values.TeamName = teamName

	// namespace, err := promptString("Namespace where your RDS is going to be accessed from?")
	// if err != nil {
	// 	return nil, err
	// }
	// values.Namespace = namespace

	return &values, nil

}
