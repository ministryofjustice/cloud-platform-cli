package environment

import (
	"os"
	"testing"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/util"
)

func cleanUpNamespacesFolder(namespace string) {
	namespaceFolder := namespaceBaseFolder + "/" + namespace
	os.RemoveAll(namespaceBaseFolder)
	os.Remove(namespaceFolder + "/resources")
	os.Remove(namespaceFolder)
	os.Remove(namespaceBaseFolder)
	os.Remove("namespaces")
}

func TestCreateNamespaceWithAnswersFile(t *testing.T) {
	answersFile := "../../testdata/environments-answers.yaml"
	err := CreateTemplateNamespace(true, answersFile)
	if err != nil {
		t.Errorf("Namespace created with answersFile errored: %s", err)
	}

	dir := namespaceBaseFolder + "/testNamespace/"
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
		namespaceFile: `name: "testNamespace"`,
	}

	for filename, searchString := range stringsInFiles {
		util.FileContainsString(t, filename, searchString)
	}

	defer cleanUpNamespacesFolder("testNamespace")
	defer os.RemoveAll(".checksum")
}

func TestReadAnswersFile(t *testing.T) {
	answersFile := "../../testdata/environments-answers.yaml"
	ns := &Namespace{}
	err := ns.readAnswersFile(answersFile)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if ns.Namespace != "testNamespace" {
		t.Errorf("Expected namespace to be testNamespace, got %s", ns.Namespace)
	}
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
		SlackChannel:          "my-team-slack_channel",
		IsProduction:          "false",
	}

	err := createNamespaceFiles(&ns)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

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
		namespaceFile:   "pod-security.kubernetes.io/enforce: restricted",
		rbacFile:        "name: \"github:my-github-team\"",
		variablesTfFile: "my-team-slack_channel",
		variablesTfFile: "my-github-team",
	}

	for filename, searchString := range stringsInFiles {
		util.FileContainsString(t, filename, searchString)
	}

	cleanUpNamespacesFolder("foobar")
}

func TestRunningOutsideEnvironmentsWorkingCopy(t *testing.T) {
	err := CreateTemplateNamespace(false, "")
	if err.Error() != "this command may only be run from within a working copy of the cloud-platform-environments repository" {
		t.Errorf("Unexpected error: %s", err)
	}
}
