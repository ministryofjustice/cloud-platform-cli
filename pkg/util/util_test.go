package util

import "testing"

// If we assign a string value to 'repository', we get it back
func TestRepository(t *testing.T) {
	re := Repository{currentRepository: "foobar"}
	str, _ := re.Repository()
	if str != "foobar" {
		t.Errorf("Something went wrong: %s", str)
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
