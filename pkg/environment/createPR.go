package environment

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os/exec"
	"strings"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/github"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/slack"
)

func createPR(description, namespace, ghToken, repo string) func(github.GithubIface, []string) (string, error) {
	b := make([]byte, 2)

	rand.Read(b) //nolint:errcheck

	fourCharUid := hex.EncodeToString(b)
	branchName := namespace + "-rds-minor-version-bump-" + fourCharUid

	return func(gh github.GithubIface, filenames []string) (string, error) {
		removeRemoteCmd := exec.Command("/bin/sh", "-c", "git remote remove origin")
		removeRemoteCmd.Start() //nolint:errcheck
		removeRemoteCmd.Wait()  //nolint:errcheck

		useGhTokenCmd := exec.Command("/bin/sh", "-c", "git remote add origin https://"+ghToken+"@github.com/ministryofjustice/"+repo)
		useGhTokenCmd.Start() //nolint:errcheck
		useGhTokenCmd.Wait()  //nolint:errcheck

		checkCmd := exec.Command("/bin/sh", "-c", "git checkout -b "+branchName)
		checkCmd.Start() //nolint:errcheck
		checkCmd.Wait()  //nolint:errcheck

		strFiles := strings.Join(filenames, " ")
		cmd := exec.Command("/bin/sh", "-c", "git add "+strFiles)
		cmd.Dir = "namespaces/live.cloud-platform.service.justice.gov.uk/" + namespace + "/resources"
		cmd.Start() //nolint:errcheck
		cmd.Wait()  //nolint:errcheck

		commitCmd := exec.Command("/bin/sh", "-c", "git commit -m 'concourse: correcting rds version drift'")
		commitCmd.Dir = "namespaces/live.cloud-platform.service.justice.gov.uk/" + namespace + "/resources"
		commitCmd.Start() //nolint:errcheck
		commitCmd.Wait()  //nolint:errcheck

		pushCmd := exec.Command("/bin/sh", "-c", "git push --set-upstream origin "+branchName)
		pushCmd.Start() //nolint:errcheck
		pushCmd.Wait()  //nolint:errcheck

		return gh.CreatePR(branchName, namespace, description)
	}
}

func postPR(prUrl, slackWebhookUrl string) {
	slackErr := slack.PostToAsk(prUrl, slackWebhookUrl)

	if slackErr != nil {
		fmt.Printf("Warning: Error posting to #ask-cloud-platform %v\n", slackErr)
	}
}
