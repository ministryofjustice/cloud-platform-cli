package environment

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
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

func fileContainsString(t *testing.T, filename string, searchString string) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if !(strings.Contains(string(contents), searchString)) {
		t.Errorf(fmt.Sprintf("Didn't find %s in contents of %s", searchString, filename))
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
	}

	_, templates := downloadAndInitialiseTemplates(ns.Namespace)

	createNamespaceFiles(templates, &ns)

	dir := namespaceBaseFolder + "/foobar/"
	namespaceFile := dir + "00-namespace.yaml"
	rbacFile := dir + "00-rbac.yaml"

	filenames := []string{
		namespaceFile,
		rbacFile,
		dir + "02-limitrange.yaml",
		dir + "03-resourcequota.yaml",
		dir + "04-networkpolicy.yaml",
		dir + "resources/main.tf",
		dir + "resources/variables.tf",
		dir + "resources/versions.tf",
	}

	for _, f := range filenames {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			t.Errorf("Expected file %s to be created", f)
		}
	}

	// test value interpolation

	type testValueInterpolation struct {
		filename string
		want     string
	}

	tests := []testValueInterpolation{
		{filename: namespaceFile, want: "name: foobar"},
		{filename: namespaceFile, want: "cloud-platform.justice.gov.uk/business-unit: \"My Biz Unit\""},
		{filename: namespaceFile, want: "cloud-platform.justice.gov.uk/environment-name: \"envname\""},
		{filename: namespaceFile, want: "cloud-platform.justice.gov.uk/application: \"My App\""},
		{filename: namespaceFile, want: "cloud-platform.justice.gov.uk/owner: \"Some Team: some-team@digital.justice.gov.uk\""},
		{filename: namespaceFile, want: "cloud-platform.justice.gov.uk/source-code: \"https://github.com/ministryofjustice/somerepo\""},
		{filename: rbacFile, want: "name: \"github:my-github-team\""},
	}

	for _, tc := range tests {
		fileContainsString(t, tc.filename, tc.want)
	}

	cleanUpNamespacesFolder("foobar")
}

func TestRunningOutsideEnvironmentsWorkingCopy(t *testing.T) {
	err := CreateTemplateNamespace(nil, nil)
	if err.Error() != "This command may only be run from within a working copy of the cloud-platform-environments repository\n" {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestOutputsInterpolatedTemplateToPath(t *testing.T) {
	filename := "fixtures/test.txt"

	templates := []*templateFromUrl{
		{
			outputPath: filename,
			content:    "name: {{ .Namespace }}",
		},
	}

	values := Namespace{
		Namespace: "mynamespace",
	}

	err := createNamespaceFiles(templates, &values)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if string(contents) != "name: mynamespace" {
		t.Errorf("Expected:\nname: mynamespace\nGot: %s\n", contents)
	}

	cleanUpNamespacesFolder("mynamespace")
	os.Remove(filename)
}
