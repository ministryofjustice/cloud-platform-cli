package environment

import (
	"os"
	"testing"
)

func TestRequireNamespaceFolder(t *testing.T) {
	// Pretend we're in the right repository
	re := RepoEnvironment{currentRepository: cloudPlatformEnvRepo}

	// Without a 00-namespace.yaml file, should fail
	err := re.mustBeInANamespaceFolder()
	if err == nil {
		t.Errorf("This should have failed")
	}

	_, err = os.Create("00-namespace.yaml")
	if err != nil {
		t.Errorf("Could not create 00-namespace.yaml")
	}

	if re.mustBeInANamespaceFolder() != nil {
		t.Errorf("This should have passed")
	}

	os.Remove("00-namespace.yaml")
}

func TestRequireCpEnvRepo(t *testing.T) {
	// Pass if we're in the right repository
	re := RepoEnvironment{currentRepository: cloudPlatformEnvRepo}
	if re.mustBeInCloudPlatformEnvironments() != nil {
		t.Errorf("This should have passed")
	}

	// Fail if we're not in the CP-env repo (which we're not)
	re = RepoEnvironment{}
	err := re.mustBeInCloudPlatformEnvironments()
	if err == nil {
		t.Errorf("This should have failed")
	}
}

// If we assign a string value to 'repository', we get it back
func TestRepoEnvironmentRepository(t *testing.T) {
	re := RepoEnvironment{currentRepository: "foobar"}
	str, _ := re.repository()
	if str != "foobar" {
		t.Errorf("Something went wrong: %s", str)
	}
}

// If we don't assign a string value to 'repository', we get whatever the
// current git repository is called
func TestRepoEnvironmentDefaultRepository(t *testing.T) {
	re := RepoEnvironment{}
	str, _ := re.repository()
	if str != "cloud-platform-cli" {
		t.Errorf("Expected cloud-platform-cli, got: x%sx", str)
	}
}
