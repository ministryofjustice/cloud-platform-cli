package environment

import (
	"io/ioutil"
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
	ns := Namespace{Namespace: "foobar"}

	_, templates := downloadAndInitialiseTemplates(ns.Namespace)
	if len(templates) != 8 {
		t.Errorf("Expected 8 templates, got: %d", len(templates))
	}

	filename := namespaceBaseFolder + "/foobar/00-namespace.yaml"
	createNamespaceFiles(templates, &ns)
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("Expected file %s to be created", filename)
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
