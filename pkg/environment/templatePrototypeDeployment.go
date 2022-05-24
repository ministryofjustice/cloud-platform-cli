package environment

import (
	"fmt"
	"os/exec"

	"github.com/MakeNowJust/heredoc"
)

const prototypeDeploymentTemplateUrl = "https://raw.githubusercontent.com/ministryofjustice/cloud-platform-terraform-github-prototype/branch-testing/templates"

func CreateDeploymentPrototype() error {

	_, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		fmt.Println("This command only runs from a git repository\n")
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

//------------------------------------------------------------------------------

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

	proto.Branch = q.value

	return &proto, nil
}

func createPrototypeDeploymentFiles(p *Prototype) error {

	ghDir := ".github/workflows/"

	copyUrlToFile(prototypeDeploymentTemplateUrl+"/cd.yaml", ghDir+"cd-"+p.Branch+".yaml")
	copyUrlToFile(prototypeDeploymentTemplateUrl+"/kubernetes-deploy.tpl", "kubernetes-deploy-"+p.Branch+".tpl")
	return nil
}
