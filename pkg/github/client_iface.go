package github

import (
	"github.com/google/go-github/github"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/util"
)

// Github for testing purposes.

//go:generate mockery --name=Github  --structname=Github --output=pkg/mocks/github --dir pkg/github

type GithubIface interface {
	ListMergedPRs(date util.Date, count int) ([]Nodes, error)
	GetChangedFiles(int) ([]*github.CommitFile, error)
	IsMerged(prNumber int) (bool, error)
	GetContents(path string) (*github.RepositoryContent, error)
}
