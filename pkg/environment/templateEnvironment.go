package environment

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

type templateEnvironment struct {
	IsProduction          bool
	Namespace             string
	Environment           string
	GithubTeam            string
	SlackChannel          string
	BusinessUnit          string
	Application           string
	InfrastructureSupport string
	TeamName              string
	SourceCode            string
	Owner                 string
	validPath             bool
}

type templateEnvironmentFile struct {
	outputPath string
	content    string
	name       string
	url        string
}

// CreateTemplateNamespace creates the terraform files from environment's template folder
func CreateTemplateNamespace(cmd *cobra.Command, args []string) error {

	templates := []*templateEnvironmentFile{
		{
			url:  "https://raw.githubusercontent.com/ministryofjustice/cloud-platform-environments/cli-template/namespace-resources-cli-template/00-namespace.yaml",
			name: "00-namespace.yaml",
		},
		{
			url:  "https://raw.githubusercontent.com/ministryofjustice/cloud-platform-environments/cli-template/namespace-resources-cli-template/01-rbac.yaml",
			name: "01-rbac.yaml",
		},
		{
			url:  "https://raw.githubusercontent.com/ministryofjustice/cloud-platform-environments/cli-template/namespace-resources-cli-template/02-limitrange.yaml",
			name: "02-limitrange.yaml",
		},
		{
			url:  "https://raw.githubusercontent.com/ministryofjustice/cloud-platform-environments/cli-template/namespace-resources-cli-template/03-resourcequota.yaml",
			name: "03-resourcequota.yaml",
		},
		{
			url:  "https://raw.githubusercontent.com/ministryofjustice/cloud-platform-environments/cli-template/namespace-resources-cli-template/04-networkpolicy.yaml",
			name: "04-networkpolicy.yaml",
		},
		{
			url:  "https://raw.githubusercontent.com/ministryofjustice/cloud-platform-environments/cli-template/namespace-resources-cli-template/resources/main.tf",
			name: "resources/main.tf",
		},
		{
			url:  "https://raw.githubusercontent.com/ministryofjustice/cloud-platform-environments/cli-template/namespace-resources-cli-template/resources/versions.tf",
			name: "resources/versions.tf",
		},
	}

	err := initTemplateNamespace(templates)
	if err != nil {
		return (err)
	}

	namespaceValues, err := templateNamespaceSetValues()
	if err != nil {
		return (err)
	}

	err = setupPaths(templates, namespaceValues.Namespace)
	if err != nil {
		return (err)
	}

	for _, i := range templates {
		t, err := template.New("namespaceTemplates").Parse(i.content)
		if err != nil {
			return err
		}

		f, err := os.Create(i.outputPath)
		if err != nil {
			return err
		}

		err = t.Execute(f, namespaceValues)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Namespace files generated under namespaces/live-1.cloud-platform.service.justice.gov.uk/%s\n", namespaceValues.Namespace)
	color.Info.Tips("Please review before raising PR")

	return nil
}

func templateNamespaceSetValues() (*templateEnvironment, error) {
	values := templateEnvironment{}

	GithubTeams, err := getGitHubTeams()
	if err != nil {
		return nil, err
	}

	Namespace := promptString{label: "Namespace name", defaultValue: ""}
	err = Namespace.promptString()
	if err != nil {
		return nil, err
	}

	Environment := promptString{label: "Environment name", defaultValue: ""}
	err = Environment.promptString()
	if err != nil {
		return nil, err
	}

	IsProduction := promptYesNo{label: "Is Production?", defaultValue: 0}
	err = IsProduction.promptyesNo()
	if err != nil {
		return nil, err
	}

	Application := promptString{label: "Application name", defaultValue: ""}
	err = Application.promptString()
	if err != nil {
		return nil, err
	}

	GithubTeam, err := promptSelectGithubTeam(GithubTeams)

	businessUnit := promptString{label: "Business Unit", defaultValue: ""}
	err = businessUnit.promptString()
	if err != nil {
		return nil, err
	}

	SlackChannel := promptString{label: "Slack Channel", defaultValue: ""}
	err = SlackChannel.promptString()
	if err != nil {
		return nil, err
	}

	teamName := promptString{label: "Team's name", defaultValue: ""}
	err = teamName.promptString()
	if err != nil {
		return nil, err
	}

	InfrastructureSupport := promptString{label: "Team's email", defaultValue: ""}
	err = InfrastructureSupport.promptString()
	if err != nil {
		return nil, err
	}

	SourceCode := promptString{label: "Source Code", defaultValue: ""}
	err = SourceCode.promptString()
	if err != nil {
		return nil, err
	}

	Owner := promptString{label: "Owner", defaultValue: ""}
	err = Owner.promptString()
	if err != nil {
		return nil, err
	}

	values.Application = Application.value
	values.BusinessUnit = businessUnit.value
	values.Namespace = Namespace.value
	values.GithubTeam = GithubTeam
	values.Environment = Environment.value
	values.IsProduction = IsProduction.value
	values.SlackChannel = SlackChannel.value
	values.InfrastructureSupport = InfrastructureSupport.value
	values.SourceCode = SourceCode.value
	values.Owner = Owner.value
	values.TeamName = teamName.value

	return &values, nil
}

func initTemplateNamespace(t []*templateEnvironmentFile) error {
	for _, s := range t {
		content, err := downloadTemplate(s.url)
		if err != nil {
			return err
		}
		s.content = content
	}

	return nil
}

func setupPaths(t []*templateEnvironmentFile, namespace string) error {
	path, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return errors.New("You are outside cloud-platform-environment repo")
	}
	fullPath := strings.TrimSpace(string(path))
	for _, s := range t {
		s.outputPath = fullPath + fmt.Sprintf("/namespaces/live-1.cloud-platform.service.justice.gov.uk/%s/", namespace) + s.name
	}

	err = os.Mkdir(fullPath+fmt.Sprintf("/namespaces/live-1.cloud-platform.service.justice.gov.uk/%s/", namespace), 0755)
	if err != nil {
		return err
	}
	err = os.Mkdir(fullPath+fmt.Sprintf("/namespaces/live-1.cloud-platform.service.justice.gov.uk/%s/resources", namespace), 0755)
	if err != nil {
		return err
	}

	return nil
}
