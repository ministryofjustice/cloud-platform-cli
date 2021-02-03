package environment

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
)

func CreateTemplatePrototype() error {
	// TODO - uncomment this block
	// re := RepoEnvironment{}
	// err := re.mustBeInCloudPlatformEnvironments()
	// if err != nil {
	// 	return err
	// }

	proto, err := promptUserForPrototypeValues()
	if err != nil {
		return (err)
	}

	err = createNamespaceFiles(&proto.Namespace)
	if err != nil {
		return err
	}

	fmt.Println("CreateTemplatePrototype...")
	fmt.Println("Name: " + proto.Namespace.Namespace)
	return nil
}

//------------------------------------------------------------------------------

func promptUserForPrototypeValues() (*Prototype, error) {
	proto := Prototype{}
	values := Namespace{}

	q := userQuestion{
		description: heredoc.Doc(`Please choose a hostname for your prototype.
			 This must consist only of lower-case letters, digits and
			 dashes.

			 This will be;
			 * the name of the prototype's namespace on the Cloud Platform
			 * the name of the prototype's github repository
			 * part of the prototype's URL on the web

			 e.g. if you choose "my-awesome-prototype", then the eventual
			 URL of the prototype will be:

			 https://my-awesome-prototype.apps.live-1.cloud-platform.service.justice.gov.uk/

			 `),
		prompt:    "Name",
		validator: new(namespaceNameValidator),
	}
	q.getAnswer()
	// TODO: check that there isn't already a namespace or github repository with this name
	values.Namespace = q.value

	q = userQuestion{
		description: heredoc.Doc(`What is the name of your GitHub team?
			The users in this GitHub team will be assigned administrator permission
			for this Cloud Platform environment, and the github repository.

			Please enter the name in lower-case, with hyphens instead of spaces
			i.e. "Check My Diary" -> "check-my-diary"

			(this must be an exact match, or you will not have access to your
			namespace or github repository)",
			 `),
		prompt:    "GitHub Team",
		validator: new(githubTeamNameValidator),
	}
	q.getAnswer()
	values.GithubTeam = q.value

	q = userQuestion{
		description: heredoc.Doc(`Which part of the MoJ is responsible for this service?
			 `),
		prompt:    "Business Unit",
		validator: new(businessUnitValidator),
	}
	q.getAnswer()
	values.BusinessUnit = q.value

	q = userQuestion{
		description: heredoc.Doc(`What is the best slack channel (without the '#')
			to use if we need to contact your team?
			(If you don't have a team slack channel, please create one)",
			 `),
		prompt:    "Team Slack Channel",
		validator: new(slackChannelValidator),
	}
	q.getAnswer()
	values.SlackChannel = q.value

	q = userQuestion{
		description: heredoc.Doc(`What is the email address for the team
			which owns the application?
			(this should not be a named individual's email address)
			 `),
		prompt:    "Team Email",
		validator: new(teamEmailValidator),
	}
	q.getAnswer()

	q = userQuestion{
		description: heredoc.Doc(`Which team in your organisation is responsible
			for this application? (e.g. Sentence Planning)
			 `),
		prompt:    "Team",
		validator: new(notEmptyValidator),
	}
	q.getAnswer()
	values.Owner = q.value

	values.InfrastructureSupport = q.value

	// We can infer all the following, for a prototype
	values.Environment = "development"
	values.IsProduction = "false"
	values.Application = "Gov.UK Prototype Kit"
	values.SourceCode = "https://github.com/ministryofjustice/" + values.Namespace

	fmt.Println(`
Prototype kit websites must be protected by HTTP basic
authentication, so that citizens don't mistake them for
real government services.
You need to choose a username and password for your site.

NB: The username and password you choose will be stored in
plaintext in a public github repository, so do not choose any
sensitive values here.`)

	q = userQuestion{
		description: heredoc.Doc(`Please choose a username for your site:
		`),
		prompt:    "Username",
		validator: new(notEmptyValidator),
	}
	q.getAnswer()
	proto.BasicAuthUsername = q.value

	q = userQuestion{
		description: heredoc.Doc(`Please choose a password for your site:
		`),
		prompt:    "Password",
		validator: new(notEmptyValidator),
	}
	q.getAnswer()
	proto.BasicAuthPassword = q.value

	proto.Namespace = values

	return &proto, nil
}
