package environment

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/github"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/slack"
)

func createPR(description, namespace, ghToken, repo string) func(github.GithubIface, []string) (string, error) {
	b := make([]byte, 2)
	if _, err := rand.Read(b); err != nil {
		return func(gh github.GithubIface, files []string) (string, error) {
			return "", fmt.Errorf("failed to generate random suffix: %w", err)
		}
	}
	fourCharUid := hex.EncodeToString(b)
	branchName := namespace + "-rds-minor-version-bump-" + fourCharUid

	return func(gh github.GithubIface, filenames []string) (string, error) {
		repoPath := "namespaces/live.cloud-platform.service.justice.gov.uk/" + namespace + "/resources"

		removeRemoteCmd := exec.Command("/bin/sh", "-c", "git remote remove origin")
		if err := removeRemoteCmd.Run(); err != nil {
			return "", fmt.Errorf("failed to remove remote origin: %w", err)
		}

		useGhTokenCmd := exec.Command("/bin/sh", "-c", "git remote add origin https://"+ghToken+"@github.com/ministryofjustice/"+repo)
		if err := useGhTokenCmd.Run(); err != nil {
			return "", fmt.Errorf("failed to remote add origin: %w", err)
		}

		pulls, err := gh.ListOpenPRs(namespace)
		if err != nil {
			log.Printf("Warning: error listing open PRs: %v", err)
		}
		if len(pulls) > 0 {
			return "", errors.New("a PR is already open for this namespace, skipping")
		}

		checkMainCmd := exec.Command("/bin/sh", "-c", "git checkout main")
		if err := checkMainCmd.Run(); err != nil {
			return "", fmt.Errorf("failed to checkout main: %w", err)
		}

		checkBranchCmd := exec.Command("/bin/sh", "-c", "git checkout -b "+branchName)
		if err := checkBranchCmd.Run(); err != nil {
			return "", fmt.Errorf("failed to create new branch: %w", err)
		}

		addCmd := exec.Command("/bin/sh", "-c", "git add "+strings.Join(filenames, " "))
		addCmd.Dir = repoPath
		if err := addCmd.Run(); err != nil {
			log.Printf("[ERROR] Failed to git add: %v", err)
			return "", fmt.Errorf("failed to git add: %w", err)
		}

		commitCmd := exec.Command("/bin/sh", "-c",
			"git -c user.name='cloud-platform-moj' -c user.email='cloudplatform@justiceuk.onmicrosoft.com' commit -m 'concourse: correcting rds version drift'",
		)
		commitCmd.Dir = repoPath
		if err := commitCmd.Run(); err != nil {
			log.Printf("[ERROR] Failed to git commit: %v", err)
			return "", fmt.Errorf("failed to git commit: %w", err)
		}

		pushCmd := exec.Command("/bin/sh", "-c", "git push --set-upstream origin "+branchName)
		pushCmd.Dir = repoPath
		if err := pushCmd.Run(); err != nil {
			log.Printf("[ERROR] Failed to push branch: %v", err)
			return "", fmt.Errorf("failed to push branch: %w", err)
		}

		prUrl, err := gh.CreatePR(branchName, namespace, description)
		if err != nil {
			log.Printf("[ERROR] Failed to create GitHub PR: %v", err)
			return "", fmt.Errorf("failed to create PR: %w", err)
		}

		return prUrl, nil
	}
}

func postPR(prUrl, slackWebhookUrl string) {
	if err := slack.PostToAsk(prUrl, slackWebhookUrl); err != nil {
		fmt.Printf("Warning: Error posting to #ask-cloud-platform: %v\n", err)
	}
}
