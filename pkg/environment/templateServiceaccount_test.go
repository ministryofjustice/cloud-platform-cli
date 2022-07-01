package environment

import (
	"os"
	"testing"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/util"
)

func TestCreateServiceAccountFile(t *testing.T) {
	filename := "resources/serviceaccount.tf"
	os.Mkdir("resources", 0o755)

	err := createSvcAccTfFile()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	moduleName := "github.com/ministryofjustice/cloud-platform-terraform-serviceaccount"
	util.FileContainsString(t, filename, moduleName)

	os.Remove(filename)
	os.Remove("resources")
}
