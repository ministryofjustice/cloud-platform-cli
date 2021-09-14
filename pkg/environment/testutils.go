package environment

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
)

func fileContainsString(t *testing.T, filename string, searchString string) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if !(strings.Contains(string(contents), searchString)) {
		t.Errorf(fmt.Sprintf("Didn't find string: %s in file: %s", searchString, filename))
	}
}

// clone takes a Git repository and it will look to create a local copy of the repo
func clone(repo, path string) error {
	_, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:      repo,
		Progress: os.Stdout,
	})
	if err != nil {
		return err
	}

	return nil
}
