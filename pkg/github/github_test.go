package githubClient

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/go-github/github"
	mocks "github.com/ministryofjustice/cloud-platform-cli/pkg/mocks/github"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/util"
	"github.com/shurcooL/githubv4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func NewTestGithubClient(token, repo, owner string, tfMock *mocks.Github) *GithubClient {
	if tfMock == nil {
		m := new(mocks.Github)
		m.On("ListMergedPRs", mock.Anything).Return(nil, nil)
		m.On("GetChangedFiles", mock.Anything, mock.Anything).Return(nil)
		tfMock = m
	}

	return &GithubClient{
		V3:         github.NewClient(nil),
		V4:         githubv4.NewClient(nil),
		Repository: repo,
		Owner:      owner,
	}
}

func TestNewGithubClient(t *testing.T) {
	type args struct {
		token string
		repo  string
		owner string
	}
	tests := []struct {
		name string
		args args
		want *GithubClient
	}{
		{
			name: "happy path",
			args: args{
				token: "testtoken",
				repo:  "testrepo",
				owner: "testowner",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			actual := NewGithubClient(tt.args.token, tt.args.repo, tt.args.owner)
			assert.NotNil(t, actual)
		})
	}
}

func TestGithubClient_ListMergedPRs(t *testing.T) {
	type args struct {
		date  util.Date
		count int
	}
	tests := []struct {
		name         string
		args         args
		GithubClient *GithubClient
		listErr      error
		expectError  bool
	}{
		{
			name: "one PR merged",
			args: args{
				date:  util.Date{},
				count: 1,
			},
			GithubClient: &GithubClient{},
			listErr:      nil,
			expectError:  false,
		},
		{
			name: "list error",
			args: args{
				date:  util.Date{},
				count: 1,
			},
			GithubClient: &GithubClient{},
			listErr:      errors.New("init error"),
			expectError:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(mocks.Github)
			m.On("ListMergedPRs", mock.Anything).Return(tt.listErr, nil)
			m.On("GetChangedFiles", mock.Anything, mock.Anything).Return(nil)
			ghClient := NewTestGithubClient("testtoken", "testrepo", "testowner", m)
			_, err := ghClient.ListMergedPRs(tt.args.date, tt.args.count)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestGithubClient_GetChangedFiles(t *testing.T) {
	type fields struct {
		V3         *github.Client
		V4         *githubv4.Client
		Repository string
		Owner      string
	}
	type args struct {
		prNumber int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*github.CommitFile
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &GithubClient{
				V3:         tt.fields.V3,
				V4:         tt.fields.V4,
				Repository: tt.fields.Repository,
				Owner:      tt.fields.Owner,
			}
			got, err := m.GetChangedFiles(tt.args.prNumber)
			if (err != nil) != tt.wantErr {
				t.Errorf("GithubClient.GetChangedFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GithubClient.GetChangedFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}
