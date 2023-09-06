package environment

import (
	"os"
	"testing"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/util"
)

func TestCreatePrototype(t *testing.T) {
	ns := Namespace{
		Namespace:             "foobar",
		BusinessUnit:          "My Biz Unit",
		Environment:           "envname",
		Application:           "My App",
		Owner:                 "Some Team",
		InfrastructureSupport: "some-team@digital.justice.gov.uk",
		SourceCode:            "https://github.com/ministryofjustice/somerepo",
		GithubTeam:            "my-github-team",
		SlackChannel:          "my-team-slack_channel",
		IsProduction:          "false",
	}
	proto := Prototype{
		Namespace:         ns,
		BasicAuthUsername: "myusername",
		BasicAuthPassword: "mypassword",
	}

	err := createPrototypeFiles(&proto)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	dir := namespaceBaseFolder + "/foobar/"
	namespaceFile := dir + "00-namespace.yaml"
	rbacFile := dir + "01-rbac.yaml"
	variablesTfFile := dir + "resources/variables.tf"
	ecrTfFile := dir + "resources/ecr.tf"

	filenames := []string{
		namespaceFile,
		rbacFile,
		dir + "02-limitrange.yaml",
		dir + "03-resourcequota.yaml",
		dir + "04-networkpolicy.yaml",
		dir + "resources/main.tf",
		variablesTfFile,
		ecrTfFile,
		dir + "resources/basic-auth.tf",
		dir + "resources/versions.tf",
	}

	for _, f := range filenames {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			t.Errorf("Expected file %s to be created", f)
		}
	}

	stringsInFiles := map[string]string{
		namespaceFile:   "name: foobar",
		namespaceFile:   "cloud-platform.justice.gov.uk/business-unit: \"My Biz Unit\"",
		namespaceFile:   "cloud-platform.justice.gov.uk/environment-name: \"envname\"",
		namespaceFile:   "cloud-platform.justice.gov.uk/application: \"My App\"",
		namespaceFile:   "cloud-platform.justice.gov.uk/owner: \"Some Team: some-team@digital.justice.gov.uk\"",
		namespaceFile:   "cloud-platform.justice.gov.uk/source-code: \"https://github.com/ministryofjustice/somerepo\"",
		namespaceFile:   "cloud-platform.justice.gov.uk/is-production: \"false\"",
		rbacFile:        "name: \"github:my-github-team\"",
		variablesTfFile: "my-team-slack_channel",
		variablesTfFile: "my-github-team",
		ecrTfFile:       "github_repositories = [var.namespace]",
	}

	for filename, searchString := range stringsInFiles {
		util.FileContainsString(t, filename, searchString)
	}

	cleanUpNamespacesFolder("foobar")
}

func TestOutsideEnvironmentsWorkingCopy(t *testing.T) {
	err := CreateTemplatePrototype()
	if err.Error() != "this command may only be run from within a working copy of the cloud-platform-environments repository" {
		t.Errorf("Unexpected error: %s", err)
	}
}
