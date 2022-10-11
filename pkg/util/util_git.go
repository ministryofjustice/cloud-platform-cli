package util

import "github.com/google/go-github/github"

type utilGit interface {
	GetChangedFiles(token, repo, owner string, prNumber int) ([]*github.CommitFile, error)
}
