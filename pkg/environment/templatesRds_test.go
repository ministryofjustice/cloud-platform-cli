package environment

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestCreatesRdsTfFile(t *testing.T) {
	filename := "resources/rds.tf"
	os.Mkdir("resources", 0755)

	err := createRdsTfFile()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("Expected file %s to be created", filename)
	}

	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	moduleName := "github.com/ministryofjustice/cloud-platform-terraform-rds-instance"
	if !(strings.Contains(string(contents), moduleName)) {
		t.Errorf("Didn't find %s in contents of %s", moduleName, filename)
	}

	os.Remove(filename)
	os.Remove("resources")
}
