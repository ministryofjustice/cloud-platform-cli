package environment

import (
	"fmt"
	"text/template"

	"github.com/gookit/color"
)

const fileName = "05-serviceaccount.yaml"
const templateLocation = "https://raw.githubusercontent.com/ministryofjustice/cloud-platform-environments/main/namespace-resources-cli-template/05-serviceaccount.yaml"

type ServiceAccount struct {
	Name      string
	Namespace string
}

// CreateTemplateServiceAccount sets and creates a template file containing all
// the necessary values to create a serviceaccount resource in Kubernetes. It
// will only execute in a directory with a namespace resource i.e. 00-namespace.yaml.
func CreateTemplateServiceAccount(name string) error {
	re := RepoEnvironment{}
	err := re.mustBeInANamespaceFolder()
	if err != nil {
		return err
	}

	v := ServiceAccount{}

	_, err = v.setSvcValues(name)
	if err != nil {
		return err
	}

	err = v.createSvcFile()
	if err != nil {
		return err
	}

	fmt.Println("Service account resource created")
	color.Info.Tips("Please review before raising PR")

	return nil
}

// setValues creates a ServiceAccount object and sets its namespace value to those inside
// a Namespace object. The name variable is passed as an argument.
func (v *ServiceAccount) setSvcValues(name string) (*ServiceAccount, error) {
	ns := Namespace{}
	ns.ReadYaml()

	v.Name = name
	v.Namespace = ns.Namespace

	return v, nil
}

// createFile uses the values of a ServiceAccount object to interpolate the serviceaccount
// template, it then creates a file in the current working directory.
func (v *ServiceAccount) createSvcFile() error {
	templateFile, err := downloadTemplate(templateLocation)
	if err != nil {
		return (err)
	}

	tpl := template.Must(template.New("").Parse(templateFile))

	f, _ := outputFileWriter(fileName)
	err = tpl.Execute(f, v)
	if err != nil {
		return nil
	}

	return nil
}