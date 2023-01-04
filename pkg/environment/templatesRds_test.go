package environment

import (
	"os"
	"testing"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/util"
)

func TestCreatesRdsTfFile(t *testing.T) {
	filename := "resources/rds-postgresql.tf"
	err := os.Mkdir("resources", 0o755)
	if err != nil {
		t.Error(err)
	}

	rdsFile, err := createRdsTfFile("postgresql")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if rdsFile != filename {
		t.Errorf("Expected file %s but got %s", filename, rdsFile)
	}

	moduleName := "github.com/ministryofjustice/cloud-platform-terraform-rds-instance"
	util.FileContainsString(t, filename, moduleName)

	os.Remove(filename)
	os.Remove("resources")
}

func TestRdsFileAlreadyExists(t *testing.T) {
	filename := "resources/rds-postgresql.tf"
	_ = os.Mkdir("resources", 0o755)

	_, err := os.Create(filename)
	if err != nil {
		t.Error(err)
	}

	_, err = createRdsTfFile("postgresql")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	os.RemoveAll("resources")
}

func TestTfFileExists(t *testing.T) {
	filename := "testFile"

	_, err := os.Create(filename)
	if err != nil {
		t.Error(err)
	}

	if !tfFileExists(filename) {
		t.Errorf("Expected %s to exist", filename)
	}

	os.RemoveAll(filename)
}
