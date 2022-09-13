package main_test

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/go-cmdtest"
)

var update = flag.Bool("update", false, "update test files with results")

func TestCmdOutput(t *testing.T) {
	if err := exec.Command("go", "build", "-o", "cloud-platform", ".").Run(); err != nil {
		t.Fatalf("Unable to build the binary %s", err)
	}
	defer os.Remove("cloud-platform")

	ts, err := cmdtest.Read("testdata")
	if err != nil {
		t.Fatal(err)
	}

	path, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	answersFile := filepath.Join(path, "testdata", "a.yaml")
	err = os.Setenv("ANSWERS_FILE", answersFile)
	if err != nil {
		t.Fatal(err)
	}

	ts.Commands["cloud-platform"] = cmdtest.Program("cloud-platform")
	ts.Run(t, *update)

	if err := testEnvironmentExists(t); err != nil {
		t.Fatalf("Unable to find the test environment %s", err)
	}
}

func testEnvironmentExists(t *testing.T) error {
	t.Helper()

	// Check file path exists
	if err := checkFilePath(t, "namespaces/live.cloud-platform.service.justice.gov.uk/testNamespace/00-namespace.yaml"); err != nil {
		return err
	}

	// defer os.RemoveAll("namespaces")
	// defer os.RemoveAll(".checksum")

	return nil
}

func checkFilePath(t *testing.T, path string) error {
	t.Helper()

	_, testFileName, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("failed get real working directory from caller")
	}

	projectRootDir := filepath.Dir(testFileName)
	filePath := filepath.Join(projectRootDir, path)
	fmt.Println(filePath)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", filePath)
	}

	return nil
}
