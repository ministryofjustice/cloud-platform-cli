package environment

import (
	"errors"
	"os/exec"
	"strings"
)

// Get information about the current execution context, wrt. the environments
// repo.  Answer questions like:
//   * Is the current directory inside a working copy of the right repository?
//   * Does this folder contain a file with a specific name (e.g. 00-namespace.yml)?
//   * What is the name of the namespace whose definition folder I am in?

type RepoEnvironment struct {
	repository string
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
