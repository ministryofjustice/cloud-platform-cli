package environment

import (
	"fmt"
	"os"
	"text/template"

	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

type ServiceAccount struct {
	Name      string
	Namespace string
}

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
