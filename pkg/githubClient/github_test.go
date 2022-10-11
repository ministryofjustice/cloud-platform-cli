package githubClient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
