package prototype

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/util"
)

const prototypeDeploymentTemplateUrl = "https://raw.githubusercontent.com/ministryofjustice/cloud-platform-environments/main/namespace-resources-cli-template/resources/prototype/templates"
const prototypeRepoUrl = "https://raw.githubusercontent.com/ministryofjustice/moj-prototype-template/main"

func CreateDeploymentPrototype(skipDockerFiles bool) error {

	// Build the url based on the repository they are in
	util := util.Repository{}

	repo, err := util.Repository()
	if err != nil {
		fmt.Println("This command only runs from a git repository")
		return err
	}

	branch, err := util.GetBranch()
	if err != nil {
		return err
	}

	err = createPrototypeDeploymentFiles(branch)
	if err != nil {
		return err
	}

	host := repo + "-" + branch

	fmt.Printf(`
Please run:

    git add ./.github/workflows/cd-%s.yaml kubernetes-deploy-%s.tpl Dockerfile .dockerignore start.sh

	git commit -m "Add docker build and deployment files"

    and push the commit to the branch.

Shortly after your pull request with the above commit is merged, you should a continuous deployment
Github Action running against the branch your prototype github repository automatically deployed to your gov.uk prototype kit website. This usually takes
around 5 minutes.

Your prototype kit website will be served at the URL:

    https://%s.apps.live.cloud-platform.service.justice.gov.uk/

If you have any questions or feedback, please post them in #ask-cloud-platform
on slack.

`, branch, branch, host)

	return nil
}

func createPrototypeDeploymentFiles(branch string) error {

	ghDir := ".github/workflows/"
	err := os.MkdirAll(ghDir, 0o755)
	if err != nil {
		return err
	}
	ghActionFile := ghDir + "cd-" + p.BranchName + ".yaml"

	copyUrlToFile(prototypeDeploymentTemplateUrl+"/cd.yaml", ghActionFile)

	copyUrlToFile(prototypeRepoUrl+"/Dockerfile", "Dockerfile")
	copyUrlToFile(prototypeRepoUrl+"/.dockerignore", ".dockerignore")
	copyUrlToFile(prototypeRepoUrl+"/start.sh", "start.sh")

	input, err := ioutil.ReadFile(ghActionFile)
	if err != nil {
		return err
	}

	output := bytes.Replace(input, []byte("branch-name"), []byte(p.BranchName), -1)

	if err = ioutil.WriteFile(ghActionFile, output, 0666); err != nil {
		return err
	}

	copyUrlToFile(prototypeDeploymentTemplateUrl+"/kubernetes-deploy.tpl", "kubernetes-deploy-"+p.BranchName+".tpl")

	return nil

}