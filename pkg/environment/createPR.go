package environment

import (
	"os/exec"
	"strings"
)

func createPR(branchName, description string, filenames []string, a *Apply) (string, error) {
	exec.Command("git", "checkout", "-b", branchName).Output()

	strFiles := strings.Join(filenames, " ")
	cmd := exec.Command("/bin/sh", "-c", "git add "+strFiles)
	cmd.Dir = "namespaces/live.cloud-platform.service.justice.gov.uk/" + a.Options.Namespace + "/resources"
	cmd.Start()
	cmd.Wait()

	commitCmd := exec.Command("/bin/sh", "-c", "git commit -m 'concourse: correcting rds version drift'")
	commitCmd.Dir = "namespaces/live.cloud-platform.service.justice.gov.uk/" + a.Options.Namespace + "/resources"
	commitCmd.Start()
	commitCmd.Wait()

	pushCmd := exec.Command("/bin/sh", "-c", "git push --set-upstream origin "+branchName)
	pushCmd.Start()
	pushCmd.Wait()

	return a.GithubClient.CreatePR(branchName, a.Options.Namespace, description)
}
