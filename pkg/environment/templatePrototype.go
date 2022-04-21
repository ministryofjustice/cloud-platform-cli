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

Shortly after your pull request is merged, you should see new files in your
github repository:

	https://github.com/ministryofjustice/%s

Files to build a docker image to run the prototype site
	Dockerfile
	.dockerignore
	start.sh

A continuous deployment (CD) workflow, targeting the Cloud Platform
	.github/workflows/cd.yaml
	kubernetes-deploy.tpl

Changes merged and pushed to the 'main' branch of your prototype github repository will be
automatically deployed to your gov.uk prototype kit website. This usually takes
around 5 minutes.

Your prototype kit website will be served at the URL:

    https://%s.apps.live.cloud-platform.service.justice.gov.uk/

If you have any questions or feedback, please post them in #ask-cloud-platform
on slack.

`, namespaceBaseFolder+"/"+s, s, s)

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

			 https://my-awesome-prototype.apps.live.cloud-platform.service.justice.gov.uk/

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
			for this Cloud Platform environment.

			Please enter the name in lower-case, with hyphens instead of spaces
			i.e. "Check My Diary" -> "check-my-diary"

			(this must be an exact match, or you will not have access to your
			namespace)",
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
		description: heredoc.Doc(`Which team in your organisation is responsible
			for this application? (e.g. Sentence Planning)
			 `),
		prompt:    "Team",
		validator: new(notEmptyValidator),
	}
	q.getAnswer()
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
See the username and password config in basic-autf.tf and
their values stored in the basic-auth namespace secret.`)

	proto.Namespace = values

	return &proto, nil
}

func createPrototypeFiles(p *Prototype) error {
	err := createNamespaceFiles(&p.Namespace)
	if err != nil {
		return err
	}

	p.appendBasicAuthVariables()

	nsdir := namespaceBaseFolder + "/" + p.Namespace.Namespace

	copyUrlToFile(prototypeTemplateUrl+"/ecr.tf", nsdir+"/resources/ecr.tf")
	copyUrlToFile(prototypeTemplateUrl+"/serviceaccount.tf", nsdir+"/resources/serviceaccount.tf")
	copyUrlToFile(prototypeTemplateUrl+"/basic-auth.tf", nsdir+"/resources/basic-auth.tf")
	copyUrlToFile(prototypeTemplateUrl+"/github-repo.tf", nsdir+"/resources/github-repo.tf")

	return nil
}
