package environment

import (
	"os"
	"testing"
)

func cleanUpPrototypeDeploymentFiles() {
	os.Remove(".github/workflows/cd-test-branch.yaml")
	os.Remove("kubernetes-deploy-test-branch.tpl")
}

func TestCreateDeploymentPrototype(t *testing.T) {

	proto := Prototype{
		BranchName: "test-branch",
	}

	createPrototypeDeploymentFiles(&proto)

	githubActionFile := "./.github/workflows/cd-test-branch.yaml"
	deploymentFile := "./kubernetes-deploy-test-branch.tpl"

	filenames := []string{
		githubActionFile,
		deploymentFile,
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
