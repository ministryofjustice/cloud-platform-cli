package release

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Release struct {
	innerStruct myRelease
	BinaryName  string
}

// These attributes need to be exported so that the json.Unmarshal call works
// correctly. But we don't want them to be exported to callers of this package,
// so we wrap them in a private, innner struct which is not exported.
type myRelease struct {
	releaseJson    []byte
	Owner          string
	RepoName       string
	CurrentVersion string
	LatestTag      string `json:"tag_name"`
}

func New(owner string, repoName string, currentVersion string, binaryName string) Release {
	innerStruct := myRelease{
		Owner:          owner,
		RepoName:       repoName,
		CurrentVersion: currentVersion,
	}
	return Release{
		BinaryName:  binaryName,
		innerStruct: innerStruct,
	}
}

func (r *Release) UpgradeIfNotLatest() {
	err, latest := r.isLatestVersion()
	if err == nil && latest {
		return
	} else if err == nil {
		err = r.informUserToUpgrade()
	}

	fmt.Println(err.Error())
	os.Exit(1)
}

// -------------------------------------------------------------

func (r *Release) isLatestVersion() (error, bool) {
	err := r.innerStruct.getLatestReleaseInfo()
	if err != nil {
		return err, false
	}

	return nil, r.innerStruct.LatestTag == r.innerStruct.CurrentVersion
}

func (r *Release) informUserToUpgrade() error {
	fmt.Printf("Update required. Current version: %s, Latest version: %s\n\n", r.innerStruct.CurrentVersion, r.innerStruct.LatestTag)
	return fmt.Errorf("to upgrade the cloud platform cli, run `brew update && brew upgrade cloud-platform-cli` or grab the latest version from https://github.com/ministryofjustice/cloud-platform-cli/releases")
}

func (r *myRelease) getLatestReleaseInfo() error {
	body, err := r.getLatestReleaseJson()
	if err != nil {
		return err
	}

	json.Unmarshal(body, r)

	return nil
}

func (r *myRelease) getLatestReleaseJson() ([]byte, error) {
	body := r.releaseJson

	if len(body) == 0 {
		response, err := http.Get(r.latestReleaseUrl())
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}

		r.releaseJson = body
	}

	return r.releaseJson, nil
}

func (r *myRelease) latestReleaseUrl() string {
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", r.Owner, r.RepoName)
}
