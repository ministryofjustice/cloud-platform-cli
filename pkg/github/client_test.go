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

func (m *mockGithub) Get(ctx context.Context, owner, repo string, number int) (*github.PullRequest, *github.Response, error) {
	return nil, nil, nil
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

type mockPullRequestService struct {
	pr    *github.PullRequest
	files []*github.CommitFile
}

func (m *mockPullRequestService) Get(ctx context.Context, owner, repo string, number int) (*github.PullRequest, *github.Response, error) {
	return m.pr, nil, nil
}

func (m *mockPullRequestService) ListFiles(ctx context.Context, owner, repo string, number int, opts *github.ListOptions) ([]*github.CommitFile, *github.Response, error) {
	return m.files, nil, nil
}

func (m *mockPullRequestService) IsMerged(ctx context.Context, owner, repo string, number int) (bool, *github.Response, error) {
	return true, nil, nil
}

func (m *mockPullRequestService) Create(ctx context.Context, owner, repo string, pr *github.NewPullRequest) (*github.PullRequest, *github.Response, error) {
	return nil, nil, nil
}

func (m *mockPullRequestService) List(ctx context.Context, owner, repo string, opts *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error) {
	return nil, nil, nil
}

func TestGithubClient_PRDetails(t *testing.T) {
	tests := []struct {
		name        string
		prNumber    int
		clusterName string
		pr          *github.PullRequest
		files       []*github.CommitFile
		want        []string
		wantErr     bool
		errContains string
	}{
		{
			name:        "successful extraction with valid namespace path",
			prNumber:    123,
			clusterName: "live",
			pr: &github.PullRequest{
				Head: &github.PullRequestBranch{
					Ref: github.Ptr("feature-branch"),
				},
			},
			files: []*github.CommitFile{
				{
					Filename: github.Ptr("namespaces/live.cloud-platform.service.justice.gov.uk/test-namespace/resources/variables.tf"),
				},
				{
					Filename: github.Ptr("other/file.txt"),
				},
			},
			want:    []string{"feature-branch", "test-namespace"},
			wantErr: false,
		},
		{
			name:        "no matching namespace path",
			prNumber:    124,
			clusterName: "live",
			pr: &github.PullRequest{
				Head: &github.PullRequestBranch{
					Ref: github.Ptr("another-branch"),
				},
			},
			files: []*github.CommitFile{
				{
					Filename: github.Ptr("other/path/file.txt"),
				},
				{
					Filename: github.Ptr("random/file.yaml"),
				},
			},
			want:    []string{"another-branch", ""},
			wantErr: false,
		},
		{
			name:        "multiple files with first matching",
			prNumber:    125,
			clusterName: "live",
			pr: &github.PullRequest{
				Head: &github.PullRequestBranch{
					Ref: github.Ptr("dev-branch"),
				},
			},
			files: []*github.CommitFile{
				{
					Filename: github.Ptr("namespaces/live.cloud-platform.service.justice.gov.uk/first-namespace/resources/00-namespace.yaml"),
				},
				{
					Filename: github.Ptr("namespaces/live.cloud-platform.service.justice.gov.uk/second-namespace/resources/main.tf"),
				},
			},
			want:    []string{"dev-branch", "first-namespace"},
			wantErr: false,
		},
		{
			name:        "path with insufficient parts",
			prNumber:    126,
			clusterName: "live",
			pr: &github.PullRequest{
				Head: &github.PullRequestBranch{
					Ref: github.Ptr("short-path-branch"),
				},
			},
			files: []*github.CommitFile{
				{
					Filename: github.Ptr("namespaces/live.cloud-platform.service.justice.gov.uk"),
				},
			},
			want:    []string{"short-path-branch", ""},
			wantErr: false,
		},
		{
			name:        "path with different cluster name",
			prNumber:    126,
			clusterName: "staging",
			pr: &github.PullRequest{
				Head: &github.PullRequestBranch{
					Ref: github.Ptr("short-path-branch"),
				},
			},
			files: []*github.CommitFile{
				{
					Filename: github.Ptr("namespaces/staging.cloud-platform.service.justice.gov.uk/test-namespace/resources/variables.tf"),
				},
			},
			want:    []string{"short-path-branch", "test-namespace"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPRService := &mockPullRequestService{
				pr:    tt.pr,
				files: tt.files,
			}

			c := &GithubClient{
				Owner:        "test-owner",
				Repository:   "test-repo",
				PullRequests: mockPRService,
			}

			got, err := c.PRDetails(context.Background(), tt.prNumber, tt.clusterName)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
