package prototype

import (
	"os"
	"testing"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/util"
)

func cleanUpPrototypeDeploymentFiles(branch string, skipDocker bool) {
	os.RemoveAll(".github")
	os.Remove(".github")
	os.Remove("kubernetes-deploy-" + branch + ".tpl")
	if !skipDocker {
		cleanUpPrototypeDockerFiles()
	}
}

func cleanUpPrototypeDockerFiles() {
	os.Remove("Dockerfile")
	os.Remove(".dockerignore")
	os.Remove("start.sh")
}

func TestCreateDeploymentPrototype(t *testing.T) {
	createPrototypeDeploymentFiles("branch-01", false)

	githubActionFile := "./.github/workflows/cd-branch-01.yaml"
	deploymentFile := "./kubernetes-deploy-branch-01.tpl"
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
		githubActionFile: "branch-01",
		deploymentFile:   "{BRANCH}",
	}

	for filename, searchString := range stringsInFiles {
		util.FileContainsString(t, filename, searchString)
	}

	cleanUpPrototypeDeploymentFiles("branch-01", false)
}

func TestCreateDeploymentPrototypeWithSkip(t *testing.T) {
	createPrototypeDeploymentFiles("branch-02", true)

	githubActionFile := "./.github/workflows/cd-branch-02.yaml"
	deploymentFile := "./kubernetes-deploy-branch-02.tpl"
	dockerFile := "./Dockerfile"
	dockerIgnoreFile := "./.dockerignore"
	startShFile := "./start.sh"

	filenames := []string{
		githubActionFile,
		deploymentFile,
	}

	filenamesSkip := []string{
		dockerFile,
		dockerIgnoreFile,
		startShFile,
	}

	for _, f := range filenames {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			t.Errorf("Expected file %s to be created", f)
		}
	}

	for _, f := range filenamesSkip {
		if _, err := os.Stat(f); os.IsExist(err) {
			t.Errorf("Expected file %s to not be created because of skip set to true", f)
		}
	}

	stringsInFiles := map[string]string{
		githubActionFile: "branch-02",
		deploymentFile:   "{BRANCH}",
	}

	for filename, searchString := range stringsInFiles {
		util.FileContainsString(t, filename, searchString)
	}

	cleanUpPrototypeDeploymentFiles("branch-02", true)
}
