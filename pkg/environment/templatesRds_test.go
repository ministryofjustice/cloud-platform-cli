package environment

import (
	"os"
	"testing"
)

func TestCreatesRdsTfFile(t *testing.T) {
	filename := "resources/rds.tf"
	os.Mkdir("resources", 0755)

	err := createRdsTfFile()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	moduleName := "github.com/ministryofjustice/cloud-platform-terraform-rds-instance"
	fileContainsString(t, filename, moduleName)

	os.Remove(filename)
	os.Remove("resources")
}
