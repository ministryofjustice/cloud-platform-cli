package release

import (
	"io/ioutil"
	"testing"
)

func TestRunningCurrentVersion(t *testing.T) {
	r := New("owner", "reponame", "9.10.11", "myapp")
	json, _ := ioutil.ReadFile("fixtures/9.10.11-version.json")
	r.innerStruct.releaseJson = json

	_, latest := r.isLatestVersion()
	if !latest {
		t.Errorf("Expected version to be latest")
	}
}

func TestNotLatest(t *testing.T) {
	r := New("owner", "reponame", "8.8.8", "myapp")
	json, _ := ioutil.ReadFile("fixtures/9.10.11-version.json")
	r.innerStruct.releaseJson = json

	_, latest := r.isLatestVersion()
	if latest {
		t.Errorf("Expected version to not be latest")
	}
}

func TestLatestReleaseUrl(t *testing.T) {
	r := New("owner", "reponame", "9.10.11", "myapp")

	url := r.innerStruct.latestReleaseUrl()
	expected := "https://api.github.com/repos/owner/reponame/releases/latest"

	if url != expected {
		t.Errorf("Expected: %s, got: %s", expected, url)
	}
}
