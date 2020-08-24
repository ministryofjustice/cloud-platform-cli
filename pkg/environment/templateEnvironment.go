package environment

import (
	"fmt"
	"os"
	"text/template"

	"github.com/MakeNowJust/heredoc"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

func CreateTemplateNamespace(cmd *cobra.Command, args []string) error {
	q := userQuestion{
		description: heredoc.Doc(`
			Is this a production namespace?
			Please enter "true" or "false"
			 `),
		prompt:    "Prouduction?",
		validator: new(trueFalseValidator),
	}
	q.getAnswer()

	fmt.Printf("You answered: x%sy\n", q.value)
	os.Exit(0)

	re := RepoEnvironment{}
	err := re.mustBeInCloudPlatformEnvironments()
	if err != nil {
		return err
	}

	nsValues, err := promptUserForNamespaceValues()
	if err != nil {
		return (err)
	}

	err = createNamespaceFiles(nsValues)
	if err != nil {
		return err
	}

	fmt.Printf("Namespace files generated under %s/%s\n", namespaceBaseFolder, nsValues.Namespace)
	color.Info.Tips("Please review before raising PR")

	return nil
}

//------------------------------------------------------------------------------

func promptUserForNamespaceValues() (*Namespace, error) {

	values := Namespace{}

	q := userQuestion{
		description: heredoc.Doc(`
			 What is the name of your namespace?
			 This should be of the form: <application>-<environment>.
			 e.g. myapp-dev (lower-case letters and dashes only)
			 `),
		prompt:    "Name",
		validator: new(namespaceNameValidator),
	}
	q.getAnswer()
	values.Namespace = q.value

	r := new(regexValidator)
	r.regex = `^[a-z]+$`
	q = userQuestion{
		description: heredoc.Doc(`
		     What type of application environment is this namespace for?
			 e.g. development, staging, production
			 `),
		prompt:    "Environment",
		validator: r,
	}
	q.getAnswer()
	values.Environment = q.value

	q = userQuestion{
		description: heredoc.Doc(`
			Is this a production namespace?
			Please enter "true" or "false"
			 `),
		prompt:    "Prouduction?",
		validator: new(trueFalseValidator),
	}
	q.getAnswer()
	values.IsProduction = q.value

	q = userQuestion{
		description: heredoc.Doc(`
		    What is the name of your application/service?
			(e.g. Send money to a prisoner)
			 `),
		prompt:    "Application",
		validator: new(notEmptyValidator),
	}
	q.getAnswer()
	values.Application = q.value

	q = userQuestion{
		description: heredoc.Doc(`
		    What is the name of your Github team?
			Please enter the name in lower-case, with hyphens instead of spaces
			i.e. "Check My Diary" -> "check-my-diary"
			(this must be an exact match, or you will not have access to your namespace)",
			 `),
		prompt:    "Github Team",
		validator: new(githubTeamNameValidator),
	}
	q.getAnswer()
	values.GithubTeam = q.value

	q = userQuestion{
		description: heredoc.Doc(`
            Which part of the MoJ is responsible for this service?
			 `),
		prompt:    "Business Unit",
		validator: new(businessUnitValidator),
	}
	q.getAnswer()
	values.BusinessUnit = q.value

	q = userQuestion{
		description: heredoc.Doc(`
		    What is the best slack channel (without the '#')
			to use if we need to contact your team?
			(If you don't have a team slack channel, please create one)",
			 `),
		prompt:    "Team Slack Channel",
		validator: new(slackChannelValidator),
	}
	q.getAnswer()
	values.SlackChannel = q.value

	q = userQuestion{
		description: heredoc.Doc(`
            What is the email address for the team
			which owns the application?
			(this should not be a named individual's email address)
			 `),
		prompt:    "Team Email",
		validator: new(slackChannelValidator),
	}
	q.getAnswer()
	values.InfrastructureSupport = q.value

	q = userQuestion{
		description: heredoc.Doc(`
            What is the Github repository URL of
			the source code for this application?
			 `),
		prompt:    "Github Repo",
		validator: new(githubUrlValidator),
	}
	q.getAnswer()
	values.SourceCode = q.value

	Owner := promptString{
		label:        "Which team in your organisation is responsible for this application? (e.g. Sentence Planning)",
		defaultValue: "",
	}
	err = Owner.promptString()
	if err != nil {
		return nil, err
	}

	values.Owner = Owner.value

	return &values, nil
}

func downloadAndInitialiseTemplates(namespace string) (error, []*templateFromUrl) {
	templates := []*templateFromUrl{
		{
			name: "00-namespace.yaml",
			url:  envTemplateLocation + "/" + "00-namespace.yaml",
		},
		{
			name: "01-rbac.yaml",
			url:  envTemplateLocation + "/" + "01-rbac.yaml",
		},
		{
			name: "02-limitrange.yaml",
			url:  envTemplateLocation + "/" + "02-limitrange.yaml",
		},
		{
			name: "03-resourcequota.yaml",
			url:  envTemplateLocation + "/" + "03-resourcequota.yaml",
		},
		{
			name: "04-networkpolicy.yaml",
			url:  envTemplateLocation + "/" + "04-networkpolicy.yaml",
		},
		{
			name: "resources/main.tf",
			url:  envTemplateLocation + "/" + "resources/main.tf",
		},
		{
			name: "resources/versions.tf",
			url:  envTemplateLocation + "/" + "resources/versions.tf",
		},
		{
			name: "resources/variables.tf",
			url:  envTemplateLocation + "/" + "resources/variables.tf",
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

func createNamespaceFiles(nsValues *Namespace) error {
	err := os.MkdirAll(fmt.Sprintf("%s/%s/resources", namespaceBaseFolder, nsValues.Namespace), 0755)
	if err != nil {
		return err
	}

	err, templates := downloadAndInitialiseTemplates(nsValues.Namespace)
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
