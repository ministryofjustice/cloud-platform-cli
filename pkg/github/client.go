package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/github"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/util"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

var _ GithubPullRequestsService = (*github.PullRequestsService)(nil)

type GithubPullRequestsService interface {
	ListFiles(ctx context.Context, owner string, repo string, number int, opt *github.ListOptions) ([]*github.CommitFile, *github.Response, error)
	IsMerged(ctx context.Context, owner string, repo string, number int) (bool, *github.Response, error)
}

// GithubClient for handling requests to the Github V3 and V4 APIs.
type GithubClient struct {
	V3           *github.Client
	V4           *githubv4.Client
	Repository   string
	Owner        string
	PullRequests GithubPullRequestsService
}

// GithubClient for handling requests to the Github V3 and V4 APIs.
type GithubClientConfig struct {
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
func NewGithubClient(config *GithubClientConfig, token string) *GithubClient {

	client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))

	v3 := github.NewClient(client)
	v4 := githubv4.NewClient(client)

	return &GithubClient{
		V3:           v3,
		V4:           v4,
		Repository:   config.Repository,
		Owner:        config.Owner,
		PullRequests: v3.PullRequests,
	}
}

// ListMergedPRs takes date and number of PRs count as input, search the github using Graphql api for
// list of PRs (title,url) between the first and last date provided
func (gh *GithubClient) ListMergedPRs(date util.Date, count int) ([]Nodes, error) {

	var query struct {
		Search struct {
			Nodes []Nodes
		} `graphql:"search(first: $count, query: $searchQuery, type: ISSUE)"`
	}
	fmt.Printf("Searching Merged PRs from %s to %s\n", date.Last, date.First)
	variables := map[string]interface{}{
		"searchQuery": githubv4.String(
			fmt.Sprintf(`repo:%s/%s is:pr is:closed merged:%s..%s`,
				gh.Owner, gh.Repository, date.Last, date.First)),
		"count": githubv4.Int(count),
	}

	err := gh.V4.Query(context.Background(), &query, variables)
	if err != nil {
		return nil, err
	}

	return query.Search.Nodes, nil
}

func (gh *GithubClient) GetChangedFiles(prNumber int) ([]*github.CommitFile, error) {
	repos, _, err := gh.PullRequests.ListFiles(
		context.Background(),
		gh.Owner,
		gh.Repository,
		prNumber,
		&github.ListOptions{PerPage: 100})
	if err != nil {
		return nil, err
	}

	return repos, nil
}

func (gh *GithubClient) IsMerged(prNumber int) (bool, error) {
	merged, _, err := gh.PullRequests.IsMerged(context.Background(), gh.Owner, gh.Repository, prNumber)
	if err != nil {
		return false, err
	}

	return merged, nil
}
