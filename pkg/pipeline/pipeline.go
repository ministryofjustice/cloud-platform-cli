package pipeline

import (
	"fmt"
	"os"
	"os/exec"
)

func runCmd(cliCmd string, args []string) {
	cmd := exec.Command(cliCmd, args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println("could not run command: ", err)
	}
}

func DeletePipelineShellCmds(clusterName string, branchName string) {
	runCmd("fly", []string{"--target", "manager", "login", "--team-name", "main", "--concourse-url", "https://concourse.cloud-platform.service.justice.gov.uk/"})
	runCmd("bash", []string{"-c", "wget -qO- https://raw.githubusercontent.com/ministryofjustice/cloud-platform-terraform-concourse/main/pipelines/manager/main/delete-cluster.yaml | fly --target manager set-pipeline -p delete-cluster -c - -v cluster_name=" + clusterName + " -v branch_name=" + branchName})
	runCmd("fly", []string{"--target", "manager", "trigger-job", "-j", "delete-cluster/delete"})
	fmt.Println("Deleting... https://concourse.cloud-platform.service.justice.gov.uk/teams/main/pipelines/delete-cluster/jobs/delete/builds/latest")
}

func CordonAndDrainPipelineShellCmds(clusterName, nodeGroup string) {
	strCmd := fmt.Sprintf("wget -qO- https://raw.githubusercontent.com/ministryofjustice/cloud-platform-terraform-concourse/main/pipelines/manager/main/cordon-and-drain-nodes.yaml | fly -t manager set-pipeline --pipeline cordon-and-drain-nodes  --config - -v node_group_to_drain=%s -v cluster_name=%s", nodeGroup, clusterName)

	runCmd("fly", []string{"--target", "manager", "login", "--team-name", "main", "--concourse-url", "https://concourse.cloud-platform.service.justice.gov.uk/"})
	runCmd("bash", []string{"-c", strCmd})
	runCmd("fly", []string{"--target", "manager", "trigger-job", "-j", "cordon-and-drain-nodes/cordon-and-drain-nodes"})
	fmt.Println("Cordoning and Draining... https://concourse.cloud-platform.service.justice.gov.uk/teams/main/pipelines/cordon-and-drain-nodes/jobs/cordon-and-drain-nodes/builds/latest")
}

func CreateCustomPipelineShellCmds(branchName string) {
	runCmd("fly", []string{"--target", "manager", "login", "--team-name", "main", "--concourse-url", "https://concourse.cloud-platform.service.justice.gov.uk/"})
	runCmd("bash", []string{"-c", "wget -qO- https://raw.githubusercontent.com/ministryofjustice/cloud-platform-terraform-concourse/main/pipelines/manager/main/custom-cluster.yaml | fly --target manager set-pipeline -p custom-cluster -c - -v branch_name=" + branchName})
	runCmd("fly", []string{"--target", "manager", "trigger-job", "-j", "custom-cluster/create"})
	fmt.Println("Creating... https://concourse.cloud-platform.service.justice.gov.uk/teams/main/pipelines/custom-cluster/jobs/create/builds/latest")
}

func TestCustomPipelineShellCmds(clusterName string, branchName string) {
	runCmd("fly", []string{"--target", "manager", "login", "--team-name", "main", "--concourse-url", "https://concourse.cloud-platform.service.justice.gov.uk/"})
	runCmd("bash", []string{"-c", "wget -qO- https://raw.githubusercontent.com/ministryofjustice/cloud-platform-terraform-concourse/main/pipelines/manager/main/custom-cluster.yaml | fly --target manager set-pipeline -p custom-cluster -c - -v cluster_name=" + clusterName + " -v branch_name=" + branchName})
	runCmd("fly", []string{"--target", "manager", "trigger-job", "-j", "custom-cluster/custom-integration-tests"})
	fmt.Println("Testing... https://concourse.cloud-platform.service.justice.gov.uk/teams/main/pipelines/custom-cluster/jobs/custom-integration-tests/builds/latest")
}
