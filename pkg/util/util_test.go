package util

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// If we assign a string value to 'repository', we get it back
func TestWorkingRepository(t *testing.T) {
	re := Repository{currentRepository: "foobar"}
	str, err := re.Repository()
	if str != "foobar" {
		t.Errorf("Something went wrong: %s", str)
	}
	if err != nil {
		t.Errorf("Expected an error, got nil")
	}
}

// If we aren't in a git repo, we get an Error
func TestRepoErrorRepository(t *testing.T) {
	// Get the current working directory
	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("Something went wrong: %s", err)
	}

	// Change to a directory which isn't a git repo
	err = os.Chdir("/tmp")
	if err != nil {
		t.Errorf("Something went wrong: %s", err)
	}
	defer os.Chdir(wd)

	re := Repository{}
	_, err = re.Repository()
	if err == nil {
		t.Errorf("Expected an error, got nil")
	}
}

// If we don't assign a string value to 'repository', we get whatever the
// current git repository is called
func TestRepoDefaultRepository(t *testing.T) {
	re := Repository{}
	str, _ := re.Repository()
	if str != "cloud-platform-cli" {
		t.Errorf("Expected cloud-platform-cli, got: x%sx", str)
	}
}

// If we assign a string value to 'branch', we get it back
func TestWorkingBranch(t *testing.T) {
	re := Repository{branch: "foobar"}
	str, err := re.GetBranch()
	if str != "foobar" {
		t.Errorf("Something went wrong: %s", str)
	}
	if err != nil {
		t.Errorf("Expected an error, got nil")
	}
}

// If we don't assign a string value to 'branch', we get whatever the
// current git branch is called
func TestRepoDefaultBranch(t *testing.T) {
	branch, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		t.Errorf("Something went wrong: %s", err)
	}

	re := Repository{}
	str, _ := re.GetBranch()
	if str != strings.Trim(string(branch), "\n") {
		t.Errorf("Expected master, got: x%sx", str)
	}
}

// If we aren't in a git repo, we get an Error
func TestRepoErrorBranch(t *testing.T) {
	// Get the current working directory
	wd, _ := os.Getwd()
	// Change to a directory which isn't a git repo
	err := os.Chdir("/tmp")
	if err != nil {
		t.Errorf("Something went wrong: %s", err)
	}
	defer os.Chdir(wd)

	re := Repository{}
	_, err = re.GetBranch()
	if err == nil {
		t.Errorf("Expected an error, got nil")
	}
}
