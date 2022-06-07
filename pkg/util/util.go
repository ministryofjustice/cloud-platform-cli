package util

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type Repository struct {
	currentRepository string
	branch            string
}

type templateFromUrl struct {
	outputPath string
	content    string
	name       string
	url        string
}

// set and return the name of the git repository which the current working
// directory is located within
func (re *Repository) Repository() (string, error) {
	// using re.repository here allows us to override this method in tests, so
	// that we can run tests regardless of the current working directory
	if re.currentRepository == "" {
		path, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
		if err != nil {
			return "", errors.New("current directory is not in a git repo working copy")
		}
		arr := strings.Split(string(path), "/")
		str := arr[len(arr)-1]
		re.currentRepository = strings.Trim(str, "\n")
	}

	return re.currentRepository, nil
}

// set and return the current branch of the git repository which is in the
// current working directory
func (re *Repository) GetBranch() (string, error) {
	// using re.getBranch here allows us to override this method in tests, so
	// that we can run tests regardless of the current working directory
	if re.branch == "" {
		// Fetch the git current branch
		branch, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
		if err != nil {
			fmt.Println("Cannot get the git branch. Check if it is a git repository")
			return "", err
		}
		re.branch = strings.Trim(string(branch), "\n")
	}
	return re.branch, nil
}
