package util

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"

	"github.com/ministryofjustice/cloud-platform-environments/pkg/authenticate"
)

type Repository struct {
	currentRepository string
	branch            string
}

// set and return the name of the git repository which the current working
// directory is located within
func (re *Repository) Repository() (string, error) {
	// using re.repository here allows us to override this method in tests, so
	// that we can run tests regardless of the current working directory
	if re.currentRepository == "" {
		path, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
		if err != nil {
			return "", errors.New("current directory is not in a git repo working copy")
		}
		arr := strings.Split(string(path), "/")
		str := arr[len(arr)-1]
		re.currentRepository = strings.Trim(str, "\n")
	}

	return re.currentRepository, nil
}

// set and return the current branch of the git repository which is in the
// current working directory
func (re *Repository) GetBranch() (string, error) {
	// using re.getBranch here allows us to override this method in tests, so
	// that we can run tests regardless of the current working directory
	if re.branch == "" {
		// Fetch the git current branch
		branch, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
		if err != nil {
			fmt.Println("Cannot get the git branch. Check if it is a git repository")
			return "", err
		}
		re.branch = strings.Trim(string(branch), "\n")
	}
	return re.branch, nil
}

// Get the latest changes form the origin remove and merge into current branch
// It is assumed the current working directory is a git repo so ensure you check before calling this method
func GetLatestGitPull() error {
	// git pull of the repo
	cmd := exec.Command("git", "pull")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return err
	}
	fmt.Println("Executing git pull")

	return nil
}

func Redacted(w io.Writer, output string) {
	re := regexp.MustCompile(`(?i)password|secret|token|key|https://hooks\.slack\.com|user|arn|ssh-rsa|clientid`)
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		if re.Match([]byte(scanner.Text())) {
			fmt.Fprintln(w, "REDACTED")
		} else {
			fmt.Fprintln(w, scanner.Text())
		}
	}
}

func ChangedInPR(token, repo, owner string, prNumber int) ([]string, error) {
	client, err := authenticate.GitHubClient(token)
	if err != nil {
		return nil, err
	}
	repos, _, err := client.PullRequests.ListFiles(context.Background(), owner, repo, prNumber, nil)
	if err != nil {
		return nil, err
	}

	var namespaceNames []string
	for _, repo := range repos {
		// namespaces filepaths are assumed to come in
		// the format: namespaces/<cluster>.cloud-platform.service.justice.gov.uk/<namespaceName>
		s := strings.Split(*repo.Filename, "/")
		namespaceNames = append(namespaceNames, s[2])
	}

	return deduplicateList(namespaceNames), nil
}

// deduplicateList will simply take a slice of strings and
// return a deduplicated version.
func deduplicateList(s []string) (list []string) {
	keys := make(map[string]bool)

	for _, entry := range s {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}

	return
}
