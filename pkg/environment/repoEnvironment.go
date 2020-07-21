package environment

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type RepoEnvironment struct {
	repository string
}

func (re *RepoEnvironment) mustBeInCloudPlatformEnvironments() error {
	err, repo := re.Repository()
	if err != nil {
		return err
	}
	if repo != cloudPlatformEnvRepo {
		return errors.New(fmt.Sprintf("This command may only be run from within a working copy of the %s repository\n", cloudPlatformEnvRepo))
	}
	return nil
}

// set and return the name of the git repository which the current working
// directory is located within
func (re *RepoEnvironment) Repository() (error, string) {
	// using re.repository here allows us to override this method in tests, so
	// that we can run tests regardless of the current working directory
	if re.repository == "" {
		path, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
		if err != nil {
			return errors.New("current directory is not in a git repo working copy"), ""
		}
		arr := strings.Split(string(path), "/")
		str := arr[len(arr)-1]
		re.repository = strings.Trim(str, "\n")
	}
	return nil, re.repository
}
