//go:build integration

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
	"github.com/ministryofjustice/cloud-platform-cli/cmd"
)

var (
	update  = flag.Bool("update", false, "update test files with results")
	logging = flag.Bool("disable-golden-logging", true, "disable the long output of the golden file test")
)

// TestCmdOutput runs a command and compares its output to the expected output. It uses the go-cmdtest package to do this.
// The testdata directory contains the "golden" files, which are the expected outputs of the commands.
func TestCmdOutput(t *testing.T) {
	if err := exec.Command("go", "build", "-o", "cloud-platform", ".").Run(); err != nil {
		t.Fatalf("Unable to build the binary %s", err)
	}
	defer os.Remove("cloud-platform")

	ts, err := cmdtest.Read("testdata")
	if err != nil {
		t.Fatal(err)
	}

	ts.DisableLogging = *logging

	path, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// The only way to get cmdtest to pickup the answers file is to set an environment variable
	answersFile := filepath.Join(path, "testdata", "environments-answers.yaml")
	err = os.Setenv("ANSWERS_FILE", answersFile)
	if err != nil {
		t.Fatal(err)
	}

	ts.Commands["cloud-platform"] = cmdtest.Program("cloud-platform")
	ts.Run(t, *update)
}

// TestEnvironmentExists executes the command "cloud-platform environment create" with an answers file and asserts an outcome.
func TestEnvironmentCreateE2E(t *testing.T) {
	_, err := cmd.ExecuteCommand(t, "environment", "create", "--skip-env-check", "--answers-file", "testdata/environments-answers.yaml", "--skip-version-check")
	if err != nil {
		t.Errorf("error executing command: %s", err)
	}

	err = testEnvironmentExists(t)
	if err != nil {
		t.Errorf("error checking environment exists: %s", err)
	}

	defer os.RemoveAll("namespaces")
	defer os.RemoveAll(".checksum")
}

func testEnvironmentExists(t *testing.T) error {
	t.Helper()

	// Check file path exists
	if err := checkFilePath(t, "namespaces/live.cloud-platform.service.justice.gov.uk/testNamespace/00-namespace.yaml"); err != nil {
		return err
	}

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

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", filePath)
	}

	return nil
}
