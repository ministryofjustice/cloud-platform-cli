package githubClient

import (
	"context"
	"fmt"

	"github.com/google/go-github/github"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/util"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// Github for testing purposes.

//go:generate mockery --name=Github  --structname=Github --output=pkg/mocks/github --dir pkg/github

type Github interface {
	ListMergedPRs(date util.Date, count int) ([]Nodes, error)
	GetChangedFiles(string, string) error
}

// GithubClient for handling requests to the Github V3 and V4 APIs.
type GithubClient struct {
	V3         *github.Client
	V4         *githubv4.Client
	Repository string
	Owner      string
}

// Nodes represents the GraphQL commit node.
// https://developer.github.com/v4/object/pullrequest/
type Nodes struct {
	PullRequest struct {
		Title githubv4.String
		Url   githubv4.String
	} `graphql:"... on PullRequest"`
}

// NewGithubClient ...
func NewGithubClient(token, repo, owner string) *GithubClient {

	client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))

	v3 := github.NewClient(client)
	v4 := githubv4.NewClient(client)

	return &GithubClient{
		V3:         v3,
		V4:         v4,
		Repository: repo,
		Owner:      owner,
	}
}

// ListMergedPRs takes date and number of PRs count as input, search the github using Graphql api for
//  list of PRs (title,url) between the first and last date provided
func (m *GithubClient) ListMergedPRs(date util.Date, count int) ([]Nodes, error) {

	var query struct {
		Search struct {
			Nodes []Nodes
		} `graphql:"search(first: $count, query: $searchQuery, type: ISSUE)"`
	}

	variables := map[string]interface{}{
		"searchQuery": githubv4.String(
			fmt.Sprintf(`repo:ministryofjustice/cloud-platform-environments is:pr is:closed merged:%s..%s`,
				date.First.Format("2006-01-02T11:00:00+00:00"), date.Last.Format("2006-01-02T11:00:00+00:00"))),
		"count": githubv4.Int(count),
	}

	err := m.V4.Query(context.Background(), &query, variables)
	if err != nil {
		return nil, err
	}

	return query.Search.Nodes, nil
}

func (m *GithubClient) GetChangedFiles(prNumber int) ([]*github.CommitFile, error) {

	repos, _, err := m.V3.PullRequests.ListFiles(context.Background(), m.Owner, m.Repository, prNumber, nil)
	if err != nil {
		return nil, err
	}

	return repos, nil
}
