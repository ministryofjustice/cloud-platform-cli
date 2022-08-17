package environment

import (
	"os"
	"testing"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/util"
)

func TestCreatesRdsTfFile(t *testing.T) {
	filename := "resources/rds.tf"
	err := os.Mkdir("resources", 0o755)
	if err != nil {
		t.Error(err)
	}

	err = createRdsTfFile()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	moduleName := "github.com/ministryofjustice/cloud-platform-terraform-rds-instance"
	util.FileContainsString(t, filename, moduleName)

	os.Remove(filename)
	os.Remove("resources")
}
