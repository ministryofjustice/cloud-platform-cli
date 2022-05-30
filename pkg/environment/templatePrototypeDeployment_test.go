package environment

import (
	"os"
	"testing"
)

func cleanUpPrototypeDeploymentFiles() {
	os.RemoveAll(".github")
	os.Remove(".github")
	os.Remove("kubernetes-deploy-test-branch.tpl")
	os.Remove("Dockerfile")
	os.Remove(".dockerignore")
	os.Remove("start.sh")
}

func TestCreateDeploymentPrototype(t *testing.T) {

	proto := Prototype{
		BranchName: "test-branch",
	}

	createPrototypeDeploymentFiles(&proto)

	githubActionFile := "./.github/workflows/cd-test-branch.yaml"
	deploymentFile := "./kubernetes-deploy-test-branch.tpl"
	dockerFile := "./Dockerfile"
	dockerIgnoreFile := "./.dockerignore"
	startShFile := "./start.sh"

	filenames := []string{
		githubActionFile,
		deploymentFile,
		dockerFile,
		dockerIgnoreFile,
		startShFile,
	}

	for _, f := range filenames {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			t.Errorf("Expected file %s to be created", f)
		}
	}

	stringsInFiles := map[string]string{

		githubActionFile: "test-branch",
		deploymentFile:   "{BRANCH}",
	}

	for filename, searchString := range stringsInFiles {
		fileContainsString(t, filename, searchString)
	}

	cleanUpPrototypeDeploymentFiles()
}
