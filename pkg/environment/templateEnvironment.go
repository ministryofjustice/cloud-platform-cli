package environment

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	dir "golang.org/x/mod/sumdb/dirhash"
)

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

	err = createNamespaceFiles(nsValues)
	if err != nil {
		return err
	}

	err = createDirHash(nsValues)
	if err != nil {
		return nil
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
	_ = q.getAnswer()
	values.Namespace = q.value

	q = userQuestion{
		description: heredoc.Doc(`
			What type of application environment is this namespace for?
			e.g. development, staging, production
			 `),
		prompt:    "Environment",
		validator: new(lowercaseStringValidator),
	}
	_ = q.getAnswer()
	values.Environment = q.value

	// If the user requests a namespace for a dev-alpha environment,
	// we need to create the namespace in the dev-alpha directory.
	if strings.ToLower(q.value) == "dev-alpha" {
		namespaceBaseFolder = devAlphaBaseDir
	}

	if q.value == "development" || q.value == "dev" {
		values.IsProduction = "false"

		q = userQuestion{
			description: heredoc.Doc(`
			Is this a sandbox/scratch namespace (e.g. for experimentation, or working through a tutorial)?
			Please enter "yes" or "no"
			 `),
			prompt:    "Sandbox?",
			validator: new(yesNoValidator),
		}
		_ = q.getAnswer()
		if q.value == "yes" {
			values.ReviewAfter = reviewAfter()
		}
	} else {
		q = userQuestion{
			description: heredoc.Doc(`
			Is this a production namespace?
			Please enter "yes" or "no"
			 `),
			prompt:    "Production?",
			validator: new(yesNoValidator),
		}
		_ = q.getAnswer()
		if q.value == "yes" {
			values.IsProduction = "true"
		} else {
			values.IsProduction = "false"
		}
	}

	q = userQuestion{
		description: heredoc.Doc(`
			What is the name of your application/service?
			(e.g. Send money to a prisoner)
			 `),
		prompt:    "Application",
		validator: new(notEmptyValidator),
	}
	_ = q.getAnswer()
	values.Application = q.value

	q = userQuestion{
		description: heredoc.Doc(`
			What is the name of your GitHub team?
			The users in this GitHub team will be assigned administrator permission for this Cloud Platform environment.
			Please enter the name in lower-case, with hyphens instead of spaces
			i.e. "Check My Diary" -> "check-my-diary"
			(this must be an exact match, or you will not have access to your namespace)",
			 `),
		prompt:    "GitHub Team",
		validator: new(githubTeamNameValidator),
	}
	_ = q.getAnswer()
	values.GithubTeam = q.value

	q = userQuestion{
		description: heredoc.Doc(`
			Which part of the MoJ is responsible for this service?
			 `),
		prompt:    "Business Unit",
		validator: new(businessUnitValidator),
	}
	_ = q.getAnswer()
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
	_ = q.getAnswer()
	values.SlackChannel = q.value

	q = userQuestion{
		description: heredoc.Doc(`
			What is the email address for the team
			which owns the application?
			(this should not be a named individual's email address)
			 `),
		prompt:    "Team Email",
		validator: new(teamEmailValidator),
	}
	_ = q.getAnswer()
	values.InfrastructureSupport = q.value

	q = userQuestion{
		description: heredoc.Doc(`
			What is the Github repository URL of
			the source code for this application?
			 `),
		prompt:    "Github Repo",
		validator: new(githubUrlValidator),
	}
	_ = q.getAnswer()
	values.SourceCode = q.value

	q = userQuestion{
		description: heredoc.Doc(`
			Which team in your organisation is responsible
			for this application? (e.g. Sentence Planning)
			 `),
		prompt:    "Team",
		validator: new(notEmptyValidator),
	}
	_ = q.getAnswer()
	values.Owner = q.value

	return &values, nil
}

func downloadAndInitialiseTemplates(namespace string) ([]*templateFromUrl, error) {
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
		return nil, err
	}

	for _, s := range templates {
		s.outputPath = fmt.Sprintf("%s/%s/", namespaceBaseFolder, namespace) + s.name
	}
	return templates, nil
}

func createNamespaceFiles(nsValues *Namespace) error {
	err := os.MkdirAll(fmt.Sprintf("%s/%s/resources", namespaceBaseFolder, nsValues.Namespace), 0o755)
	if err != nil {
		return err
	}

	templates, err := downloadAndInitialiseTemplates(nsValues.Namespace)
	if err != nil {
		return err
	}

	err = createFilesFromTemplates(templates, *nsValues)
	if err != nil {
		return err
	}

	return nil
}

// createDirHash calls the dirhash package to create a sha256 hash of the users
// namespace directory. This value is written to a file at the root of the
// cloud-platform-environments repository.
func createDirHash(nsValues *Namespace) error {
	// A DefaultHash is a required argument in the dirhash package
	var DefaultHash dir.Hash = dir.Hash1

	fileName := ".checksum"
	nsDir := namespaceBaseFolder + "/" + nsValues.Namespace

	hashDir, err := dir.HashDir(nsDir, nsValues.Namespace, DefaultHash)
	if err != nil {
		return err
	}

	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString("#This file is used by the auto pr github action. Please commit" + "\n")
	if err != nil {
		return err
	}

	_, err = f.WriteString(nsValues.Namespace + "\n" + hashDir + "\n")
	if err != nil {
		return err
	}

	return nil
}

func reviewAfter() string {
	return string(time.Now().AddDate(0, 3, 0).Format("2006-01-02"))
}
