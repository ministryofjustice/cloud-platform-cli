package github

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/go-github/v68/github"
	"github.com/jferrl/go-githubauth"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/util"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

var _ GithubPullRequestsService = (*github.PullRequestsService)(nil)

type GithubPullRequestsService interface {
	ListFiles(ctx context.Context, owner string, repo string, number int, opt *github.ListOptions) ([]*github.CommitFile, *github.Response, error)
	IsMerged(ctx context.Context, owner string, repo string, number int) (bool, *github.Response, error)
	Create(ctx context.Context, owner string, repo string, pr *github.NewPullRequest) (*github.PullRequest, *github.Response, error)
	List(ctx context.Context, owner string, repo string, opts *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error)
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

func NewGihubAppClient(config *GithubClientConfig, key, appid, installid string) *GithubClient {
	privateKey := []byte(key)

	appIDInt, err := strconv.ParseInt(appid, 10, 64)
	if err != nil {
		fmt.Printf("[NewGihubAppClient] Failed to parse appid '%s': %v\n", appid, err)
		return nil
	}

	installIDInt, err := strconv.ParseInt(installid, 10, 64)
	if err != nil {
		fmt.Printf("[NewGihubAppClient] Failed to parse installid '%s': %v\n", installid, err)
		return nil
	}

	appTokenSource, err := githubauth.NewApplicationTokenSource(appIDInt, privateKey)
	if err != nil {
		fmt.Printf("[NewGihubAppClient] Failed to create ApplicationTokenSource: %v\n", err)
		return nil
	}

	installationTokenSource := githubauth.NewInstallationTokenSource(installIDInt, appTokenSource)

	oauthHttpClient := oauth2.NewClient(context.Background(), installationTokenSource)

	v3 := github.NewClient(oauthHttpClient)
	v4 := githubv4.NewClient(oauthHttpClient)

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

func (gh *GithubClient) CreatePR(branchName, namespace, description string) (string, error) {
	newPR := &github.NewPullRequest{
		Title:               github.String("Fix: rds version mismatch in " + namespace),
		Head:                github.String("ministryofjustice:" + branchName),
		Base:                github.String("main"),
		Body:                github.String(description),
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := gh.PullRequests.Create(context.TODO(), "ministryofjustice", "cloud-platform-environments", newPR)
	if err != nil {
		return "", err
	}

	fmt.Printf("PR created: %s\n", pr.GetHTMLURL())
	return pr.GetHTMLURL(), nil
}

func (gh *GithubClient) ListOpenPRs(namespace string) ([]*github.PullRequest, error) {
	listOpts := &github.ListOptions{PerPage: 100, Page: 0}
	opts := &github.PullRequestListOptions{
		State:       "open",
		ListOptions: *listOpts,
	}
	allOpenPrs := []*github.PullRequest{}
	matchedOpenPRs := []*github.PullRequest{}
	paginate := 0

	prs, resp, err := gh.PullRequests.List(context.TODO(), "ministryofjustice", "cloud-platform-environments", opts)
	if err != nil {
		return nil, err
	}

	allOpenPrs = append(allOpenPrs, prs...)

	paginate = resp.NextPage

	for paginate > 0 {
		opts.ListOptions.Page = resp.NextPage
		prs, resp, err := gh.PullRequests.List(context.TODO(), "ministryofjustice", "cloud-platform-environments", opts)
		if err != nil {
			return nil, err
		}

		allOpenPrs = append(allOpenPrs, prs...)

		paginate = resp.NextPage
	}

	for _, pr := range allOpenPrs {
		if strings.Contains(*pr.Title, "Fix: rds version mismatch in "+namespace) {
			matchedOpenPRs = append(matchedOpenPRs, pr)
		}
	}

	return matchedOpenPRs, nil
}

func (gh *GithubClient) CreateComment(prNumber int, body string) error {
	comment := &github.IssueComment{
		Body: github.String(body),
	}

	_, _, err := gh.V3.Issues.CreateComment(
		context.TODO(),
		"ministryofjustice",
		"cloud-platform-environments",
		prNumber,
		comment,
	)

	return err
}
