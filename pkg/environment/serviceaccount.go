package environment

import (
	"fmt"
	"os"
	"text/template"

	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

const fileName = "05-serviceaccount.yaml"
const templateLocation = "https://raw.githubusercontent.com/ministryofjustice/cloud-platform-environments/cli-svc/namespace-resources-cli-template/05-serviceaccount.yaml"
type ServiceAccount struct {
	Name      string
	Namespace string
}

// CreateTemplateServiceAccount sets and creates a template file containing all
// the necessary values to create a serviceaccount resource in Kubernetes. It
// will only execute in a directory with a namespace resource i.e. 00-namespace.yaml.
func CreateTemplateServiceAccount(cmd *cobra.Command, args []string) error {
	re := RepoEnvironment{}
	err := re.mustBeInANamespaceFolder()
	if err != nil {
		return err
	}

	values, err := promptForValues()
	if err != nil {
		return err
	}

	err = createServiceAccountFiles(values)
	if err != nil {
		return err
	}

	fmt.Printf("Serviceaccount generated under %s/%s\n", namespaceBaseFolder, values.Namespace)
	color.Info.Tips("Please review before raising PR")

	return nil
}

func promptForValues() (*ServiceAccount, error) {
// setValues creates a ServiceAccount object and sets its namespace value to those inside
// a Namespace object. It will also set the object name to "randomString" unless
// specified otherwise.
	values := ServiceAccount{}

	ServiceAccountName := promptString{
		label:        "What name would you like to call your serviceaccount? This should be lowercase e.g. circleci",
		defaultValue: "",
		validation:   "no-spaces-and-no-uppercase",
	}
	err := ServiceAccountName.promptString()
	if err != nil {
		return nil, err
	}

	Namespace := promptString{
		label:        "What is the name of your namespace? This should be of the form: <application>-<environment>. e.g. myapp-dev (lower-case letters and dashes only)",
		defaultValue: "",
		validation:   "no-spaces-and-no-uppercase",
	}
	err = Namespace.promptString()
	if err != nil {
		return nil, err
	}

	values.Namespace = Namespace.value
	values.Name = ServiceAccountName.value

	return &values, nil
}

func downloadAndInitialiseTemplate(namespace string) (error, []*templateFromUrl) {
// initialiseTemplate creates a templateFromUrl object and downloads the contents of the
// templateLocation variable, returning a slice of templateFromUrl objects. The downloadTemplateContents
// function requires a slice, so even though we're only passing one template it needs to be added to a slice.
	template := []*templateFromUrl{
		{
			name:       "05-serviceaccount.yaml",
			url:        "https://raw.githubusercontent.com/ministryofjustice/cloud-platform-environments/cli-svc/namespace-resources-cli-template/05-serviceaccount.yaml",
			outputPath: "05-serviceaccount.yaml",
		},
	}

	err := downloadTemplateContents(template)
	if err != nil {
		return err, nil
	}

	return nil, template
}

func createServiceAccountFiles(values *ServiceAccount) error {
	if _, err := os.Stat(namespaceBaseFolder + values.Namespace); os.IsNotExist(err) {
		os.Mkdir(namespaceBaseFolder+values.Namespace, 0755)
	}

	err, templates := downloadAndInitialiseTemplate(values.Namespace)
// createFile creates the template file in the current directory. As the initialiseTemplate function
// returns a slice of templateFromUrl objects, it has to be iterated over in a for loop.
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

		err = t.Execute(f, values)
		if err != nil {
			return err
		}
	}
	return nil
}
