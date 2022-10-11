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
	"time"

	"github.com/google/go-github/github"
	"github.com/ministryofjustice/cloud-platform-environments/pkg/authenticate"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type Repository struct {
	currentRepository string
	branch            string
}

type nodes struct {
	PullRequest struct {
		Title githubv4.String
		Url   githubv4.String
	} `graphql:"... on PullRequest"`
}
type Date struct {
	first time.Time
	last  time.Time
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

// NSChangedInPR get the list of changed files for a given PR. checks if the namespaces exists in the given cluster
// folder and return the list of namespaces.
func NSChangedInPR(cluster, token, repo, owner string, prNumber int) ([]string, error) {
	repos, err := getChangedFiles(token, repo, owner, prNumber)
	if err != nil {
		return nil, err
	}

	var namespaceNames []string
	for _, repo := range repos {
		// namespaces filepaths are assumed to come in
		// the format: namespaces/<cluster>.cloud-platform.service.justice.gov.uk/<namespaceName>
		s := strings.Split(*repo.Filename, "/")
		//only get namespaces from the folder that belong to the given cluster and
		// ignore changes outside namespace directories
		if s[1] == cluster {
			namespaceNames = append(namespaceNames, s[2])
		}

	}

	return deduplicateList(namespaceNames), nil
}

func getChangedFiles(token, repo, owner string, prNumber int) ([]*github.CommitFile, error) {

	client, err := authenticate.GitHubClient(token)
	if err != nil {
		return nil, err
	}
	repos, _, err := client.PullRequests.ListFiles(context.Background(), owner, repo, prNumber, nil)
	if err != nil {
		return nil, err
	}

	return repos, nil
}

// GetMergedPRs takes date and number of PRs count as input, search the github using Graphql api for
//  list of PRs (title,url) between the first and last date provided
func GetMergedPRs(date Date, count int, token string) ([]nodes, error) {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewEnterpriseClient("https://api.github.com/graphql", httpClient)

	var query struct {
		Search struct {
			Nodes []nodes
		} `graphql:"search(first: $count, query: $searchQuery, type: ISSUE)"`
	}

	variables := map[string]interface{}{
		"searchQuery": githubv4.String(
			fmt.Sprintf(`repo:ministryofjustice/cloud-platform-environments is:pr is:closed merged:%s..%s`,
				date.first.Format("2006-01-02T11:00:00+00:00"), date.last.Format("2006-01-02T11:00:00+00:00"))),
		"count": githubv4.Int(count),
	}

	err := client.Query(context.Background(), &query, variables)
	if err != nil {
		return nil, err
	}

	return query.Search.Nodes, nil
}

// GetDateLastMinute returns the current date and date - minute as Date object
func GetDateLastMinute() Date {
	var d Date
	d.first = time.Now()
	d.last = time.Now().Add(-time.Minute * 1)
	return d
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
