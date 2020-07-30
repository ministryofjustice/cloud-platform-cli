package environment

import (
	"fmt"
	"text/template"

	"github.com/gookit/color"
)

const svcAccFileName = "05-serviceaccount.yaml"

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

	err = v.createSvcAccFile(name)
	if err != nil {
		return err
	}

	fmt.Println(svcAccFileName, "created")
	color.Info.Tips("Please review before raising PR")

	return nil
}

// createSvcAccountFile uses the values of a ServiceAccount object to interpolate the serviceaccount
// template, it then creates a file in the current working directory.
func (v *ServiceAccount) createSvcAccFile(name string) error {
	err := v.setSvcAccValues(name)
	if err != nil {
		return err
	}

	tmpl, err := setSvcAccTemplate()
	if err != nil {
		return err
	}

	err = v.writeSvcAccFile(tmpl)
	if err != nil {
		return err
	}

	return nil
}

// setSvcValues takes a ServiceAccount object and populates it with the namespace
// value (from the 00-namespace.yaml file) and the name argument from cobra.
func (v *ServiceAccount) setSvcAccValues(name string) error {
	ns := Namespace{}
	ns.ReadYaml()

	v.Name = name
	v.Namespace = ns.Namespace

	return nil
}

// setSvcTemplate downloads the required template from the environments repository
// and returns it.
func setSvcAccTemplate() (string, error) {
	templateFile, err := downloadTemplate(envTemplateLocation + "/" + svcAccFileName)
	if err != nil {
		return "An error occurred", err
	}

	return templateFile, nil
}

// writeSvcAccFile uses the template returned by setSvcTemplate and writes a
// file to the current working directory.
func (v *ServiceAccount) writeSvcAccFile(tmpl string) error {
	tpl := template.Must(template.New("").Parse(tmpl))

	f, _ := outputFileWriter(svcAccFileName)
	err := tpl.Execute(f, v)
	if err != nil {
		return err
	}

	return nil
}
