package environment

import (
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

const NamespaceYamlFile = "00-namespace.yaml"

type Namespace struct {
	Application           string
	BusinessUnit          string
	Environment           string
	GithubTeam            string
	InfrastructureSupport string
	IsProduction          string
	Name                  string
	Namespace             string
	Owner                 string
	OwnerEmail            string
	SlackChannel          string
	SourceCode            string
}

func (ns *Namespace) readYaml() error {
	return ns.readYamlFile(NamespaceYamlFile)
}

// This is a public function so that we can use it in our tests
func (ns *Namespace) readYamlFile(filename string) error {
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
				IsProduction string `yaml:"cloud-platform.justice.gov.uk/is-production"`
				Environment  string `yaml:"cloud-platform.justice.gov.uk/environment-name"`
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

	ns.Name = t.Metadata.Name
	ns.IsProduction = t.Metadata.Labels.IsProduction
	ns.BusinessUnit = t.Metadata.Annotations.BusinessUnit
	ns.Owner = t.Metadata.Annotations.Owner
	ns.Environment = t.Metadata.Labels.Environment
	ns.OwnerEmail = strings.Split(t.Metadata.Annotations.Owner, ": ")[1]
	ns.Application = t.Metadata.Annotations.Application
	ns.SourceCode = t.Metadata.Annotations.SourceCode
	ns.Namespace = t.Metadata.Name

	return nil
}
