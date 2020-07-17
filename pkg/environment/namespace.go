package environment

import (
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

const NamespaceYamlFile = "00-namespace.yaml"

type Namespace struct {
	name            string
	isProduction    string
	businessUnit    string
	owner           string
	environmentName string
	ownerEmail      string
	application     string
	sourceCode      string
	namespace       string
}

func (ns *Namespace) ReadYaml() error {
	return ns.ReadYamlFile(NamespaceYamlFile)
}

// This is a public function so that we can use it in our tests
func (ns *Namespace) ReadYamlFile(filename string) error {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Failed to read namespace YAML file: %s", filename)
		return err
	}
	ns.parseYaml(contents)
	return nil
}

func (ns *Namespace) parseYaml(yamlData []byte) error {
	type envNamespace struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
		Metadata   struct {
			Name   string `yaml:"name"`
			Labels struct {
				IsProduction    string `yaml:"cloud-platform.justice.gov.uk/is-production"`
				EnvironmentName string `yaml:"cloud-platform.justice.gov.uk/environment-name"`
			} `yaml:"labels"`
			Annotations struct {
				BusinessUnit string `yaml:"cloud-platform.justice.gov.uk/business-unit"`
				Application  string `yaml:"cloud-platform.justice.gov.uk/application"`
				Owner        string `yaml:"cloud-platform.justice.gov.uk/owner"`
				SourceCode   string `yaml:"cloud-platform.justice.gov.uk/source-code"`
			} `yaml:"annotations"`
		} `yaml:"metadata"`
	}

	t := envNamespace{}

	err := yaml.Unmarshal(yamlData, &t)
	if err != nil {
		fmt.Printf("Could not decode namespace YAML: %v", err)
		return err
	}

	ns.name = t.Metadata.Name
	ns.isProduction = t.Metadata.Labels.IsProduction
	ns.businessUnit = t.Metadata.Annotations.BusinessUnit
	ns.owner = t.Metadata.Annotations.Owner
	ns.environmentName = t.Metadata.Labels.EnvironmentName
	ns.ownerEmail = strings.Split(t.Metadata.Annotations.Owner, ": ")[1]
	ns.application = t.Metadata.Annotations.Application
	ns.sourceCode = t.Metadata.Annotations.SourceCode
	ns.namespace = t.Metadata.Name

	return nil
}
