package environment

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/MakeNowJust/heredoc"
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
	proto, err := promptUserForPrototypeDeployValues()
	if err != nil {
		return (err)
	}

	err = createPrototypeDeploymentFiles(proto)
	if err != nil {
		return err
	}

	return nil
}

func promptUserForPrototypeDeployValues() (*Prototype, error) {
	proto := Prototype{}

	q := userQuestion{
		description: heredoc.Doc(`What is the branch name you want to deploy the prototype?
		e.g. app-testing
			 `),
		prompt:    "Branch name",
		validator: new(notMainBranchValidator),
	}
	q.getAnswer()

	proto.BranchName = q.value

	return &proto, nil
}

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
