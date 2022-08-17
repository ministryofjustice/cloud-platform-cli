package environment

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type RepoEnvironment struct {
	currentRepository string
}

func (re *RepoEnvironment) mustBeInCloudPlatformEnvironments() error {
	repo, err := re.repository()
	if err != nil {
		return err
	}
	if repo != cloudPlatformEnvRepo {
		return fmt.Errorf("this command may only be run from within a working copy of the %s repository", cloudPlatformEnvRepo)
	}
	return nil
}

func (re *RepoEnvironment) mustBeInANamespaceFolder() error {
	err := re.mustBeInCloudPlatformEnvironments()
	if err != nil {
		return err
	}

	if _, err := os.Stat("00-namespace.yaml"); os.IsNotExist(err) {
		return fmt.Errorf("this command may only be run from within a namespace folder in the the %s repository", cloudPlatformEnvRepo)
	}

	return nil
}

// getNamespaceName ensure we are inside namespace folder and also returns the
// namespace name
func (re *RepoEnvironment) getNamespaceName() (string, error) {
	nsFullPath, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return filepath.Base(nsFullPath), nil
}

// set and return the name of the git repository which the current working
// directory is located within
func (re *RepoEnvironment) repository() (string, error) {
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
