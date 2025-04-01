package environment

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/github"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/slack"
)

func createPR(description, namespace, ghToken, repo string) func(github.GithubIface, []string) (string, error) {
	b := make([]byte, 2)
	rand.Read(b)
	fourCharUid := hex.EncodeToString(b)
	branchName := namespace + "-rds-minor-version-bump-" + fourCharUid

	return func(gh github.GithubIface, filenames []string) (string, error) {
		repoPath := "namespaces/live.cloud-platform.service.justice.gov.uk/" + namespace + "/resources"
		log.Printf("🌿 Creating new branch: %s", branchName)

		pulls, err := gh.ListOpenPRs(namespace)
		if err != nil {
			log.Printf("⚠️ Warning: Error listing open PRs: %v\n", err)
		}
		if len(pulls) > 0 {
			return "", errors.New("a PR is already open for this namespace, skipping")
		}

		exec.Command("/bin/sh", "-c", "git checkout main && git pull").Run()
		exec.Command("/bin/sh", "-c", "git checkout -b "+branchName).Run()

		originalDir, _ := os.Getwd()
		os.Chdir(repoPath)
		defer os.Chdir(originalDir)

		log.Printf("📄 Files to stage for commit (relative): %v", filenames)
		exec.Command("/bin/sh", "-c", "git add "+strings.Join(filenames, " ")).Run()
		exec.Command("/bin/sh", "-c", "git -c user.name='cloud-platform-moj' -c user.email='platforms+githubuser@digital.justice.gov.uk' commit -m 'concourse: correcting rds version drift'").Run()
		os.Chdir(originalDir)
		exec.Command("/bin/sh", "-c", "git push --set-upstream origin "+branchName).Run()

		prUrl, err := gh.CreatePR(branchName, namespace, description)
		if err != nil {
			return "", err
		}

		log.Printf("🧹 Resetting Git state after PR")
		exec.Command("/bin/sh", "-c", "git reset --hard").Run()
		exec.Command("/bin/sh", "-c", "git clean -fd").Run()

		return prUrl, nil
	}
}

func postPR(prUrl, slackWebhookUrl string) {
	slackErr := slack.PostToAsk(prUrl, slackWebhookUrl)

	if slackErr != nil {
		fmt.Printf("Warning: Error posting to #ask-cloud-platform %v\n", slackErr)
	}
}
