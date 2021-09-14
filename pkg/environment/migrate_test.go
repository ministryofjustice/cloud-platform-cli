package environment

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGrepFile(t *testing.T) {
	hasBusinessUnit := grepFile("fixtures/foobar-namespace.yml", []byte("cloud-platform.justice.gov.uk/business-unit"))
	if hasBusinessUnit == 0 {
		t.Errorf("Business Unit annotation exist inside fixures file, grepFile() returned %v - expected: 1", hasBusinessUnit)
	}

	hasRandomAnnotation := grepFile("fixtures/foobar-namespace.yml", []byte("whatever"))
	if hasRandomAnnotation != 0 {
		t.Errorf("whatever annotation DOES NOT exist inside fixures file, grepFile() returned %v - expected: 0", hasRandomAnnotation)
	}
}

func TestMigrate(t *testing.T) {
	repoLocalPath := "./tmp/cloud-platform-environments"
	repo := "https://github.com/ministryofjustice/cloud-platform-environments.git"

	err := clone(repo, repoLocalPath)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	namespace := fmt.Sprintf("%s/namespaces/live-1.cloud-platform.service.justice.gov.uk/abundant-namespace-dev/", repoLocalPath)

	if err := os.Chdir(namespace); err != nil {
		t.Fatalf("changing working directory failed: %v", err)
	}

	err = Migrate(true)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	_, filename, _, _ := runtime.Caller(0)
	if err := os.Chdir(filepath.Dir(filename)); err != nil {
		t.Fatalf("changing working directory failed: %v", err)
	}

	err = os.RemoveAll("./tmp/")
	if err != nil {
		t.Errorf("Unexpected error deleting tmp/: %s", err)
	}

}
