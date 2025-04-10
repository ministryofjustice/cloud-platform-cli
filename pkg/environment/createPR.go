package environment

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/github"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/slack"
)

func createPR(description, namespace, ghToken, repo string) func(github.GithubIface, []string) (string, error) {
	b := make([]byte, 2)
	if _, err := rand.Read(b); err != nil {
		return func(github.GithubIface, []string) (string, error) {
			return "", fmt.Errorf("failed to generate random suffix: %w", err)
		}
	}
	fourCharUid := hex.EncodeToString(b)
	branchName := namespace + "-rds-minor-version-bump-" + fourCharUid

	return func(gh github.GithubIface, filenames []string) (string, error) {
		repoPath := "namespaces/live.cloud-platform.service.justice.gov.uk/" + namespace + "/resources"
		log.Printf("Creating new branch: %s", branchName)

		pulls, err := gh.ListOpenPRs(namespace)
		if err != nil {
			log.Printf("Warning: Error listing open PRs: %v", err)
		}
		if len(pulls) > 0 {
			return "", errors.New("a PR is already open for this namespace, skipping")
		}

		if err := exec.Command("git", "checkout", "main").Run(); err != nil {
			return "", fmt.Errorf("failed to checkout main: %w", err)
		}
		if err := exec.Command("git", "pull").Run(); err != nil {
			return "", fmt.Errorf("failed to pull main: %w", err)
		}
		if err := exec.Command("git", "checkout", "-b", branchName).Run(); err != nil {
			return "", fmt.Errorf("failed to create new branch: %w", err)
		}

		originalDir, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current dir: %w", err)
		}
		if err := os.Chdir(repoPath); err != nil {
			return "", fmt.Errorf("failed to cd to %s: %w", repoPath, err)
		}
		defer func() {
			_ = os.Chdir(originalDir)
		}()

		log.Printf("Files to stage for commit (relative): %v", filenames)
		args := append([]string{"add"}, filenames...)
		if err := exec.Command("git", args...).Run(); err != nil {
			return "", fmt.Errorf("failed to git add: %w", err)
		}

		commitMsg := "concourse: correcting rds version drift"
		commitCmd := exec.Command("git", "-c", "user.name=cloud-platform-moj", "-c", "user.email=platforms+githubuser@digital.justice.gov.uk", "commit", "-m", commitMsg)
		if err := commitCmd.Run(); err != nil {
			return "", fmt.Errorf("failed to git commit: %w", err)
		}

		if err := os.Chdir(originalDir); err != nil {
			return "", fmt.Errorf("failed to return to original dir: %w", err)
		}
		if err := exec.Command("git", "push", "--set-upstream", "origin", branchName).Run(); err != nil {
			return "", fmt.Errorf("failed to push branch: %w", err)
		}

		prUrl, err := gh.CreatePR(branchName, namespace, description)
		if err != nil {
			return "", fmt.Errorf("failed to create PR: %w", err)
		}

		return prUrl, nil
	}
}

func postPR(prUrl, slackWebhookUrl string) {
	slackErr := slack.PostToAsk(prUrl, slackWebhookUrl)
	if slackErr != nil {
		fmt.Printf("Warning: Error posting to #ask-cloud-platform %v\n", slackErr)
	}
}
