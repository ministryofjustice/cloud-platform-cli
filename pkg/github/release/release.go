package release

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"
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
		err = r.selfUpgrade()
	}

	fmt.Printf(err.Error())
	os.Exit(1)
}

// -------------------------------------------------------------

func (r *Release) isLatestVersion() (error, bool) {
	err := r.innerStruct.getLatestReleaseInfo() // TODO: memoize this
	if err != nil {
		return err, false
	}

	return nil, r.innerStruct.LatestTag == r.innerStruct.CurrentVersion
}

func (r *Release) selfUpgrade() error {
	fmt.Printf("Update required. Current version: %s, Latest version: %s\n\n", r.innerStruct.CurrentVersion, r.innerStruct.LatestTag)

	// download tarball of latest release
	tempFilePath := "/tmp/" + r.innerStruct.tarballFilename()

	fmt.Printf("Downloading latest tarball...\n  %s\n", r.innerStruct.latestTarballUrl())
	r.innerStruct.downloadFile(tempFilePath, r.innerStruct.latestTarballUrl())

	fmt.Println("Unpacking...")
	cmd := exec.Command("tar", "xzf", tempFilePath, "--cd", "/tmp/")
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error: ", err)
		return err
	}

	// move unpacked binary into place
	filename, _ := os.Executable()
	fmt.Printf("Replacing %s\n\n", filename)
	cmd = exec.Command("mv", fmt.Sprintf("/tmp/%s", r.BinaryName), filename)
	err = cmd.Run()
	if err != nil {
		return err
	}

	return errors.New("Upgrade successful. Please repeat your previous command.\n")
}

// -------------------------------------------------------------

func (r *myRelease) getLatestReleaseInfo() error {
	err, body := r.getLatestReleaseJson()
	if err != nil {
		return err
	}

	json.Unmarshal(body, r)

	return nil
}

func (r *myRelease) getLatestReleaseJson() (error, []byte) {
	body := r.releaseJson

	if len(body) == 0 {
		response, err := http.Get(r.latestReleaseUrl())
		if err != nil {
			return err, nil
		}
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err, nil
		}

		r.releaseJson = body
	}

	return nil, r.releaseJson
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func (r *myRelease) downloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func (r *myRelease) tarballFilename() string {
	return r.RepoName + "_" + r.LatestTag + "_" + runtime.GOOS + "_" + runtime.GOARCH + ".tar.gz"
}

func (r *myRelease) latestTarballUrl() string {
	return fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", r.Owner, r.RepoName, r.LatestTag, r.tarballFilename())
}

func (r *myRelease) latestReleaseUrl() string {
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", r.Owner, r.RepoName)
}
