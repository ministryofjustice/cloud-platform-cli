package release

import (
	"io/ioutil"
	"runtime"
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

func TestTarballFilename(t *testing.T) {
	r := New("owner", "reponame", "9.10.11", "myapp")
	json, _ := ioutil.ReadFile("fixtures/9.10.11-version.json")
	r.innerStruct.releaseJson = json
	r.innerStruct.getLatestReleaseInfo()

	tarball := r.innerStruct.tarballFilename()

	var expected string

	if runtime.GOOS == "darwin" {
		expected = "reponame_9.10.11_darwin_amd64.tar.gz"
	} else {
		expected = "reponame_9.10.11_linux_amd64.tar.gz"
	}

	if tarball != expected {
		t.Errorf("Expected: %s, got: %s", expected, tarball)
	}
}

func TestLatestTarballUrl(t *testing.T) {
	r := New("owner", "reponame", "9.10.11", "myapp")
	json, _ := ioutil.ReadFile("fixtures/9.10.11-version.json")
	r.innerStruct.releaseJson = json
	r.innerStruct.getLatestReleaseInfo()

	url := r.innerStruct.latestTarballUrl()

	var expected string

	if runtime.GOOS == "darwin" {
		expected = "https://github.com/owner/reponame/releases/download/9.10.11/reponame_9.10.11_darwin_amd64.tar.gz"
	} else {
		expected = "https://github.com/owner/reponame/releases/download/9.10.11/reponame_9.10.11_linux_amd64.tar.gz"
	}

	if url != expected {
		t.Errorf("Expected: %s, got: %s", expected, url)
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
