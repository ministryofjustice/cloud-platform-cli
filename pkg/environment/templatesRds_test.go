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

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("Expected file %s to be created", filename)
	}

	// TODO: test that the file contains the string
	// "github.com/ministryofjustice/cloud-platform-terraform-rds-instance"

	os.Remove(filename)
	os.Remove("resources")
}
