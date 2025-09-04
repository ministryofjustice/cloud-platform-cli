package github

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-github/v74/github"
	"github.com/stretchr/testify/assert"
)

type mockGithub struct {
	resp   []*github.CommitFile
	merged bool
}

func (m *mockGithub) ListFiles(ctx context.Context, owner string, repo string, number int, opt *github.ListOptions) ([]*github.CommitFile, *github.Response, error) {
	return m.resp, nil, nil
}

func (m *mockGithub) IsMerged(ctx context.Context, owner string, repo string, number int) (bool, *github.Response, error) {
	return true, nil, nil
}

func (m *mockGithub) Create(ctx context.Context, owner string, repo string, pr *github.NewPullRequest) (*github.PullRequest, *github.Response, error) {
	return nil, nil, nil
}

func (m *mockGithub) List(ctx context.Context, owner, repo string, opts *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error) {
	return nil, nil, nil
}

func TestNewGithubClient(t *testing.T) {
	type args struct {
		config *GithubClientConfig
		token  string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "happy path",
			args: args{
				&GithubClientConfig{
					Repository: "testrepo",
					Owner:      "testowner",
				},
				"testtoken",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := NewGithubClient(tt.args.config, tt.args.token)
			assert.NotNil(t, actual)
		})
	}
}

func TestGithubClient_GetChangedFiles(t *testing.T) {
	mc := &mockGithub{
		resp: []*github.CommitFile{
			{
				SHA:       github.String("6dcb09b5b57875f334f61aebed695e2e4193db5e"),
				Filename:  github.String("/namespaces/testctx/ns1/file1.txt"),
				Additions: github.Int(103),
				Deletions: github.Int(21),
				Changes:   github.Int(124),
				Status:    github.String("added"),
				Patch:     github.String("@@ -132,7 +132,7 @@ module Test @@ -1000,7 +1000,7 @@ module Test"),
			},
			{
				SHA:       github.String("f61aebed695e2e4193db5e6dcb09b5b57875f334"),
				Filename:  github.String("/namespaces/testctx/ns1/file2.txt"),
				Additions: github.Int(5),
				Deletions: github.Int(3),
				Changes:   github.Int(103),
				Status:    github.String("modified"),
				Patch:     github.String("@@ -132,7 +132,7 @@ module Test @@ -1000,7 +1000,7 @@ module Test"),
			},
		},
	}
	gh := &GithubClient{
		PullRequests: mc,
	}
	got, err := gh.GetChangedFiles(8344)
	if err != nil {
		t.Errorf("GithubClient.GetChangedFiles() error = %v, wantErr %v", err, nil)
		return
	}
	if !reflect.DeepEqual(got, mc.resp) {
		t.Errorf("GithubClient.GetChangedFiles() = %v, want %v", got, mc.resp)
	}
}

func TestGithubClient_IsMerged(t *testing.T) {
	mc := &mockGithub{
		merged: true,
	}
	gh := &GithubClient{
		PullRequests: mc,
	}
	got, err := gh.IsMerged(8344)
	if err != nil {
		t.Errorf("GithubClient.IsMerged() error = %v, wantErr %v", err, nil)
		return
	}
	if got != mc.merged {
		t.Errorf("GithubClient.IsMerged() = %v, want %v", got, true)
	}
}
