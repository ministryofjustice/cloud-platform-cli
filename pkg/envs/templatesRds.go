package envs

import (
	"fmt"
	"io/ioutil"
	"net/http"
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

	// getEnvironmentsFromGithub()

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

	isProduction, err := promptyesNo("Is Production?")
	if err != nil {
		return nil, err
	}
	values.IsProduction = isProduction

	environmentName, err := promptString("Environment name for RDS?")
	if err != nil {
		return nil, err
	}
	values.Environment = environmentName

	application, err := promptString("Application name?")
	if err != nil {
		return nil, err
	}
	values.Application = application

	businessUnit, err := promptString("Business Unit?")
	if err != nil {
		return nil, err
	}
	values.BusinessUnit = businessUnit

	teamName, err := promptString("Application name?")
	if err != nil {
		return nil, err
	}
	values.TeamName = teamName

	namespace, err := promptString("Namespace where your RDS is going to be accessed from?")
	if err != nil {
		return nil, err
	}
	values.Namespace = namespace

	return &values, nil

}

func getEnvironmentsFromGithub() {
	response, err := http.Get("https://api.github.com/repos/ministryofjustice/cloud-platform-environments/contents/namespaces/live-1.cloud-platform.service.justice.gov.uk")
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(data))
	}

	// jsonData := map[string]string{"firstname": "Nic", "lastname": "Raboy"}
	// jsonValue, _ := json.Marshal(jsonData)
	// fmt.Println(jsonValue)

}
