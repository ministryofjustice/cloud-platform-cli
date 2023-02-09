package environment

import (
	"errors"
	"net/http"
	"os"
)

// Helloworld takes a bool value to specify if you want to run this in a git repo.
// It will init a deploy and Dockerfile in your current directory, and return an error if it fails.
func Helloworld(ignoreRepository bool) (string, error) {
	// check if your in a git repo
	// create Dockerfile and deploydir
	// create makefile
	const githubURL = "https://raw.githubusercontent.com/ministryofjustice/cloud-platform-helloworld-ruby-app/main/kubectl_deploy"

	if !ignoreRepository && !isInGitRepo() {
		return "", errors.New("please re-run this command in a git repository.")
	}

	err := createDeploy(githubURL)
	if err != nil {
		return "", err
	}

	return "Hello World", nil
}

func createDeploy(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	// Create the file
	out, err := os.Create("./deploy")
	if err != nil {
		return err
	}

	defer out.Close()

	return nil
}

func isInGitRepo() bool {
	_, err := os.Stat(".git")
	if err == nil {
		return true
	}
	return false
}
