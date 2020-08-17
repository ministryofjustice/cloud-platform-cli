package environment

import (
	"os"
	"testing"
)

func cleanUpNamespacesFolder(namespace string) {
	namespaceFolder := namespaceBaseFolder + "/" + namespace
	os.RemoveAll(namespaceBaseFolder)
	os.Remove(namespaceFolder + "/resources")
	os.Remove(namespaceFolder)
	os.Remove(namespaceBaseFolder)
	os.Remove("namespaces")
}

func TestCreateNamespace(t *testing.T) {
	ns := Namespace{
		Namespace:             "foobar",
		BusinessUnit:          "My Biz Unit",
		Environment:           "envname",
		Application:           "My App",
		Owner:                 "Some Team",
		InfrastructureSupport: "some-team@digital.justice.gov.uk",
		SourceCode:            "https://github.com/ministryofjustice/somerepo",
		GithubTeam:            "my-github-team",
		SlackChannel:          "my-team-slack-channel",
		IsProduction:          "false",
	}

	createNamespaceFiles(&ns)

	dir := namespaceBaseFolder + "/foobar/"
	namespaceFile := dir + "00-namespace.yaml"
	rbacFile := dir + "01-rbac.yaml"
	variablesTfFile := dir + "resources/variables.tf"

	filenames := []string{
		namespaceFile,
		rbacFile,
		dir + "02-limitrange.yaml",
		dir + "03-resourcequota.yaml",
		dir + "04-networkpolicy.yaml",
		dir + "resources/main.tf",
		variablesTfFile,
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
		variablesTfFile: "my-team-slack-channel",
		variablesTfFile: "my-github-team",
	}

	for filename, searchString := range stringsInFiles {
		fileContainsString(t, filename, searchString)
	}

	cleanUpNamespacesFolder("foobar")
}

func TestRunningOutsideEnvironmentsWorkingCopy(t *testing.T) {
	err := CreateTemplateNamespace(nil, nil)
	if err.Error() != "This command may only be run from within a working copy of the cloud-platform-environments repository\n" {
		t.Errorf("Unexpected error: %s", err)
	}
}
