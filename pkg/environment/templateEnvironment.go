package environment

import (
	"fmt"
	"os"
	"text/template"

	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

// all yaml and terraform templates will be pulled from URL endpoints below here
const templatesBaseUrl = "https://raw.githubusercontent.com/ministryofjustice/cloud-platform-environments/main/namespace-resources-cli-template"

// CreateTemplateNamespace creates the terraform files from environment's template folder
func CreateTemplateNamespace(cmd *cobra.Command, args []string) error {
	re := RepoEnvironment{}
	err := re.mustBeInCloudPlatformEnvironments()
	if err != nil {
		return err
	}

	nsValues, err := promptUserForNamespaceValues()
	if err != nil {
		return (err)
	}

	err, templates := downloadAndInitialiseTemplates(nsValues.Namespace)
	if err != nil {
		return err
	}

	err = createNamespaceFiles(templates, nsValues)
	if err != nil {
		return err
	}

	fmt.Printf("Namespace files generated under %s/%s\n", namespaceBaseFolder, nsValues.Namespace)
	color.Info.Tips("Please review before raising PR")

	return nil
}

func promptUserForNamespaceValues() (*Namespace, error) {

	values := Namespace{}

	Namespace := promptString{
		label:        "What is the name of your namespace? This should be of the form: <application>-<environment>. e.g. myapp-dev (lower-case letters and dashes only)",
		defaultValue: "",
		validation:   "no-spaces-and-no-uppercase",
	}
	err := Namespace.promptString()
	if err != nil {
		return nil, err
	}

	Environment := promptString{
		label:        "What type of application environment is this namespace for? e.g. development, staging, production",
		defaultValue: "",
		validation:   "no-spaces-and-no-uppercase",
	}
	err = Environment.promptString()
	if err != nil {
		return nil, err
	}

	IsProduction := promptTrueFalse{
		label:        "Is this a production namespace? (choose 'true' or 'false')",
		defaultValue: "false",
	}
	err = IsProduction.prompttrueFalse()
	if err != nil {
		return nil, err
	}

	Application := promptString{
		label:        "What is the name of your application/service? (e.g. Send money to a prisoner)",
		defaultValue: "",
	}
	err = Application.promptString()
	if err != nil {
		return nil, err
	}

	GithubTeam := promptString{
		label:        "What is the name of your Github team? (this must be an exact match, or you will not have access to your namespace)",
		defaultValue: "",
	}
	err = GithubTeam.promptString()
	if err != nil {
		return nil, err
	}

	businessUnit := promptString{
		label:        "Which part of the MoJ is responsible for this service? (e.g HMPPS, Legal Aid Agency)",
		defaultValue: "",
	}
	err = businessUnit.promptString()
	if err != nil {
		return nil, err
	}

	SlackChannel := promptString{
		label:        "What is the best slack channel (without the '#') to use if we need to contact your team?\n(If you don't have a team slack channel, please create one)",
		defaultValue: "",
	}
	err = SlackChannel.promptString()
	if err != nil {
		return nil, err
	}

	InfrastructureSupport := promptString{
		label:        "What is the email address for the team which owns the application? (this should not be a named individual's email address)",
		defaultValue: "",
		validation:   "email",
	}
	err = InfrastructureSupport.promptString()
	if err != nil {
		return nil, err
	}

	SourceCode := promptString{
		label:        "What is the Github repository URL of the source code for this application?",
		defaultValue: "",
		validation:   "url",
	}
	err = SourceCode.promptString()
	if err != nil {
		return nil, err
	}

	Owner := promptString{
		label:        "Which team in your organisation is responsible for this application? (e.g. Sentence Planning)",
		defaultValue: "",
	}
	err = Owner.promptString()
	if err != nil {
		return nil, err
	}

	values.Application = Application.value
	values.BusinessUnit = businessUnit.value
	values.Namespace = Namespace.value
	values.GithubTeam = GithubTeam.value
	values.Environment = Environment.value
	values.IsProduction = IsProduction.value
	values.SlackChannel = SlackChannel.value
	values.InfrastructureSupport = InfrastructureSupport.value
	values.SourceCode = SourceCode.value
	values.Owner = Owner.value

	return &values, nil
}

func downloadAndInitialiseTemplates(namespace string) (error, []*templateFromUrl) {
	templates := []*templateFromUrl{
		{
			name: "00-namespace.yaml",
			url:  templatesBaseUrl + "/" + "00-namespace.yaml",
		},
		{
			name: "01-rbac.yaml",
			url:  templatesBaseUrl + "/" + "01-rbac.yaml",
		},
		{
			name: "02-limitrange.yaml",
			url:  templatesBaseUrl + "/" + "02-limitrange.yaml",
		},
		{
			name: "03-resourcequota.yaml",
			url:  templatesBaseUrl + "/" + "03-resourcequota.yaml",
		},
		{
			name: "04-networkpolicy.yaml",
			url:  templatesBaseUrl + "/" + "04-networkpolicy.yaml",
		},
		{
			name: "resources/main.tf",
			url:  templatesBaseUrl + "/" + "resources/main.tf",
		},
		{
			name: "resources/versions.tf",
			url:  templatesBaseUrl + "/" + "resources/versions.tf",
		},
		{
			name: "resources/variables.tf",
			url:  templatesBaseUrl + "/" + "resources/variables.tf",
		},
	}

	err := downloadTemplateContents(templates)
	if err != nil {
		return err, nil
	}

	for _, s := range templates {
		s.outputPath = fmt.Sprintf("%s/%s/", namespaceBaseFolder, namespace) + s.name
	}
	return nil, templates
}

func createNamespaceFiles(templates []*templateFromUrl, nsValues *Namespace) error {
	err := os.MkdirAll(fmt.Sprintf("%s/%s/resources", namespaceBaseFolder, nsValues.Namespace), 0755)
	if err != nil {
		return err
	}

	for _, i := range templates {
		t, err := template.New("").Parse(i.content)
		if err != nil {
			return err
		}

		f, err := os.Create(i.outputPath)
		if err != nil {
			return err
		}

		err = t.Execute(f, nsValues)
		if err != nil {
			return err
		}
	}
	return nil
}
