package environment

import (
	"testing"
)

func TestRequireCpEnvRepo(t *testing.T) {
	// Pass if we're in the right repository
	re := RepoEnvironment{repository: CloudPlatformEnvRepo}
	if re.MustBeInCloudPlatformEnvironments() != nil {
		t.Errorf("This should have passed")
	}

	// Fail if we're not in the CP-env repo (which we're not)
	re = RepoEnvironment{}
	err := re.MustBeInCloudPlatformEnvironments()
	if err == nil {
		t.Errorf("This should have failed")
	}
}

// If we assign a string value to 'repository', we get it back
func TestRepoEnvironmentRepository(t *testing.T) {
	re := RepoEnvironment{repository: "foobar"}
	_, str := re.Repository()
	if str != "foobar" {
		t.Errorf("Something went wrong: %s", str)
	}
}

// If we don't assign a string value to 'repository', we get whatever the
// current git repository is called
func TestRepoEnvironmentDefaultRepository(t *testing.T) {
	re := RepoEnvironment{}
	_, str := re.Repository()
	if str != "cloud-platform-cli" {
		t.Errorf("Expected cloud-platform-cli, got: x%sx", str)
	}
}
