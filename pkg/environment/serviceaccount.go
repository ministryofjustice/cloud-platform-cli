package environment

import (
	"fmt"
	"text/template"

	"github.com/gookit/color"
)

// TODO: Change cli-svc back to main
const fileName = "05-serviceaccount.yaml"
const templateLocation = "https://raw.githubusercontent.com/ministryofjustice/cloud-platform-environments/cli-svc/namespace-resources-cli-template/05-serviceaccount.yaml"

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

	values, err := setValues(name)
	if err != nil {
		return err
	}

	err = createFile(values)
	if err != nil {
		return err
	}

	fmt.Printf("Serviceaccount generated under %s/%s\n", namespaceBaseFolder, values.Namespace)
	color.Info.Tips("Please review before raising PR")

	return nil
}

// setValues creates a ServiceAccount object and sets its namespace value to those inside
// a Namespace object. The name variable is passed as an argument.
func setValues(name string) (*ServiceAccount, error) {
	values := ServiceAccount{}

	ns := Namespace{}
	ns.ReadYaml()

	values.Name = name
	values.Namespace = ns.Namespace

	return &values, nil
}

// createFile uses the values of a ServiceAccount object to interpolate the serviceaccount
// template and creates a file in the current working directory.
func createFile(values *ServiceAccount) error {
	templateFile, err := downloadTemplate(templateLocation)
	if err != nil {
		return (err)
	}

	tpl := template.Must(template.New("").Parse(templateFile))

	f, _ := outputFileWriter(fileName)
	err = tpl.Execute(f, values)
	if err != nil {
		return nil
	}

	return nil
}
