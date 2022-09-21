package environment

import (
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
)

const prototypeTemplateUrl = "https://raw.githubusercontent.com/ministryofjustice/cloud-platform-environments/main/namespace-resources-cli-template/resources/prototype"

func CreateTemplatePrototype() error {
	re := RepoEnvironment{}
	err := re.mustBeInCloudPlatformEnvironments()
	if err != nil {
		return err
	}

	proto, err := promptUserForPrototypeValues()
	if err != nil {
		return (err)
	}

	err = createPrototypeFiles(proto)
	if err != nil {
		return err
	}

	s := proto.Namespace.Namespace

	fmt.Printf(`
Please run:

    git add %s

...and raise a pull request.

If you have any questions or feedback, please post them in #ask-cloud-platform
on slack.

`, namespaceBaseFolder+"/"+s)

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

			 This should be;
			 * the name of your existing prototype's github repository

			 This will be;
			 * the name of the prototype's namespace on the Cloud Platform
			 * part of the prototype's URL on the web

			 e.g. if you choose "my-awesome-prototype", then the eventual
			 URL of the prototype will be:

			 https://my-awesome-prototype-main.apps.live.cloud-platform.service.justice.gov.uk/

			 where "main" is the branch you have published
			 `),
		prompt:    "Name",
		validator: new(namespaceNameValidator),
	}
	_ = q.getAnswer()
	// TODO: check that there isn't already a namespace or github repository with this name
	values.Namespace = q.value

	q = userQuestion{
		description: heredoc.Doc(`What is the name of your GitHub team?
			The users in this GitHub team will be assigned administrator permission
			for this Cloud Platform environment.

			Please enter the name in lower-case, with hyphens instead of spaces
			i.e. "Check My Diary" -> "check-my-diary"

			(this must be an exact match, or you will not have access to your
			namespace)",
			 `),
		prompt:    "GitHub Team",
		validator: new(githubTeamNameValidator),
	}
	_ = q.getAnswer()
	values.GithubTeam = q.value

	q = userQuestion{
		description: heredoc.Doc(`Which part of the MoJ is responsible for this service?
			 `),
		prompt:    "Business Unit",
		validator: new(businessUnitValidator),
	}
	_ = q.getAnswer()
	values.BusinessUnit = q.value

	q = userQuestion{
		description: heredoc.Doc(`What is the best slack channel (without the '#')
			to use if we need to contact your team?
			(If you don't have a team slack channel, please create one)",
			 `),
		prompt:    "Team Slack Channel",
		validator: new(slackChannelValidator),
	}
	_ = q.getAnswer()
	values.SlackChannel = q.value

	q = userQuestion{
		description: heredoc.Doc(`Which team in your organisation is responsible
			for this application? (e.g. Sentence Planning)
			 `),
		prompt:    "Team",
		validator: new(notEmptyValidator),
	}
	_ = q.getAnswer()
	values.Owner = q.value

	// We can infer all the following, for a prototype
	values.InfrastructureSupport = "platforms@digital.justice.gov.uk"
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
	_ = q.getAnswer()
	proto.BasicAuthUsername = q.value

	q = userQuestion{
		description: heredoc.Doc(`Please choose a password for your site:
		`),
		prompt:    "Password",
		validator: new(notEmptyValidator),
	}
	_ = q.getAnswer()
	proto.BasicAuthPassword = q.value

	proto.Namespace = values

	return &proto, nil
}

func createPrototypeFiles(p *Prototype) error {
	err := createNamespaceFiles(&p.Namespace)
	if err != nil {
		return err
	}

	err = p.appendBasicAuthVariables()
	if err != nil {
		return err
	}

	nsdir := namespaceBaseFolder + "/" + p.Namespace.Namespace

	err = CopyUrlToFile(prototypeTemplateUrl+"/ecr.tf", nsdir+"/resources/ecr.tf")
	if err != nil {
		return err
	}
	err = CopyUrlToFile(prototypeTemplateUrl+"/serviceaccount.tf", nsdir+"/resources/serviceaccount.tf")
	if err != nil {
		return err
	}
	err = CopyUrlToFile(prototypeTemplateUrl+"/basic-auth.tf", nsdir+"/resources/basic-auth.tf")
	if err != nil {
		return err
	}

	return nil
}

// Append the extra terraform variables required by a prototype site
// to the namespace's resources/variables.tf file
func (p *Prototype) appendBasicAuthVariables() error {
	str := `
## Prototype kit variables

variable "basic-auth-username" {
  description = "Basic auth. username of the deployed prototype website"
  default     = "` + p.BasicAuthUsername + `"
}

variable "basic-auth-password" {
  description = "Basic auth. password of the deployed prototype website"
  default     = "` + p.BasicAuthPassword + `"
}
`
	varTf := fmt.Sprintf("%s/%s/resources/variables.tf", namespaceBaseFolder, p.Namespace.Namespace)

	file, err := os.OpenFile(varTf, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(str); err != nil {
		return err
	}

	return nil
}
