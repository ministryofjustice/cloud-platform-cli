package environment

import (
	"os"
	"testing"
)

func TestCreateServiceAccountFile(t *testing.T) {
	filename := "resources/serviceaccount.tf"
	os.Mkdir("resources", 0755)

	err := createSvcAccTfFile()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	moduleName := "github.com/ministryofjustice/cloud-platform-terraform-serviceaccount"
	fileContainsString(t, filename, moduleName)

	os.Remove(filename)
	os.Remove("resources")
}
