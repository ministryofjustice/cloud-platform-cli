package terraform

import (
	"bufio"
	"os"
	"path/filepath"

	"github.com/davecgh/go-spew/spew"
)

type BulkActions struct {
	ChangedFilesPath string
}

func (b *BulkActions) Plan() error {
	dirs, _ := b.targetDirs()

	spew.Dump(dirs)

	return nil
}

func (b *BulkActions) targetDirs() ([]string, error) {
	var dirs []string // Directories where tf plan is going to be executed

	f, err := os.Open(b.ChangedFilesPath)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if contains(dirs, filepath.Dir(scanner.Text())) != true {
			dirs = append(dirs, filepath.Dir(scanner.Text()))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return dirs, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
