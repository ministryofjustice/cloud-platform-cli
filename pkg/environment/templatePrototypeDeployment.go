package environment

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

type templateGHAction struct {
	BranchName string
}

const prototypeDeploymentTemplateUrl = "https://raw.githubusercontent.com/ministryofjustice/cloud-platform-terraform-github-prototype/branch-testing/templates"

func CreateDeploymentPrototype() error {

	_, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		fmt.Println("This command only runs from a git repository")
		return err
	}

	proto := Prototype{}

	branch, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		fmt.Println("Cannot get the git branch")
		return err
	}
	proto.BranchName = strings.Trim(string(branch), "\n")

	err = createPrototypeDeploymentFiles(&proto)
	if err != nil {
		return err
	}

	re := RepoEnvironment{}

	err, repo := re.repository()
	if err != nil {
		return err
	}

	host := repo + "-" + proto.BranchName

	fmt.Printf(`
Please run:

    git add ./.github/workflows/cd-%s.yaml kubernetes-deploy-%s.tpl

...and raise a pull request.

Shortly after your pull request is merged, you should a continuous deployment
Github Action running against the branch your prototype github repository automatically deployed to your gov.uk prototype kit website. This usually takes
around 5 minutes.

Your prototype kit website will be served at the URL:

    https://%s.apps.live.cloud-platform.service.justice.gov.uk/

If you have any questions or feedback, please post them in #ask-cloud-platform
on slack.

`, proto.BranchName, proto.BranchName, host)

	return nil
}

// func promptUserForPrototypeDeployValues() (*Prototype, error) {
// 	proto := Prototype{}

// 	q := userQuestion{
// 		description: heredoc.Doc(`What is the branch name you want to deploy the prototype?
// 		e.g. app-testing
// 			 `),
// 		prompt:    "Branch name",
// 		validator: new(notMainBranchValidator),
// 	}
// 	q.getAnswer()

// 	proto.BranchName = q.value

// 	return &proto, nil
// }

func createPrototypeDeploymentFiles(p *Prototype) error {

	ghDir := ".github/workflows/"
	err := os.MkdirAll(ghDir, 0o755)
	if err != nil {
		return err
	}
	ghActionFile := ghDir + "cd-" + p.BranchName + ".yaml"

	copyUrlToFile(prototypeDeploymentTemplateUrl+"/cd.yaml", ghActionFile)

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
