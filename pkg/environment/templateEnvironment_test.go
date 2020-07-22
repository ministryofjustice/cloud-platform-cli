package environment

import (
	"io/ioutil"
	"os"
	"testing"
)

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

	values := environmentValues{
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

	os.Remove(filename)
}
