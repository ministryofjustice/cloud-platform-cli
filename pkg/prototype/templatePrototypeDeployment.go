package prototype

import (
	"bytes"
	"fmt"
	"os"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/environment"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/util"
)

const (
	prototypeResourcesUrl = "https://raw.githubusercontent.com/ministryofjustice/cloud-platform-environments/prototype-basic-auth/namespace-resources-cli-template/resources/prototype"
)

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

	err = createPrototypeDeploymentFiles(branch, skipDockerFiles)
	if err != nil {
		return err
	}

	host := repo + "-" + branch

	fmt.Printf(`
Please run:

    git add ./.github/workflows/cd-%s.yaml kubernetes-deploy-%s.tpl

	if --skip-docker-files flag is not used, then run,

	git add Dockerfile .dockerignore start.sh

	git commit -m "Add deployment files"

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

func createPrototypeDeploymentFiles(branch string, skipDockerFiles bool) error {
	if !skipDockerFiles {
		err := environment.CopyUrlToFile(prototypeResourcesUrl+"/build"+"/Dockerfile", "Dockerfile")
		if err != nil {
			return err
		}
		err = environment.CopyUrlToFile(prototypeResourcesUrl+"/build"+"/.dockerignore", ".dockerignore")
		if err != nil {
			return err
		}
		err = environment.CopyUrlToFile(prototypeResourcesUrl+"/build"+"/start.sh", "start.sh")
		if err != nil {
			return err
		}
	}

	ghDir := ".github/workflows/"
	err := os.MkdirAll(ghDir, 0o755)
	if err != nil {
		return err
	}
	ghActionFile := ghDir + "cd-" + branch + ".yaml"

	err = environment.CopyUrlToFile(prototypeResourcesUrl+"/templates"+"/cd.yaml", ghActionFile)
	if err != nil {
		return err
	}

	input, err := os.ReadFile(ghActionFile)
	if err != nil {
		return err
	}

	output := bytes.Replace(input, []byte("branch-name"), []byte(branch), -1)

	if err = os.WriteFile(ghActionFile, output, 0o666); err != nil {
		return err
	}

	err = environment.CopyUrlToFile(prototypeResourcesUrl+"/templates"+"/kubernetes-deploy.tpl", "kubernetes-deploy-"+branch+".tpl")
	if err != nil {
		return err
	}

	return nil
}
