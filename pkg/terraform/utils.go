package terraform

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

type BulkActions struct {
	ChangedFilesPath string
}

func (b *BulkActions) Plan(c *Commander) error {
	dirs, _ := b.targetDirs()

	kops := []string{"terraform/cloud-platform-components", "terraform/cloud-platform"}
	// eks := []string{"terraform/cloud-platform-eks/components", "terraform/cloud-platform-eks"}

	dirsToPlan, found := Find(dirs, kops)

	if found {
		for _, dir := range dirsToPlan {
			fmt.Println(dir)
			c.cmdDir = dir
			err := c.Plan()
			if err != nil {
				fmt.Println(err)
				return err
			}

		}
	}

	// spew.Dump(Find(dirs, kops))

	return nil
}

// Find takes a slice and looks for an element in it. If found it will
// return it's key, otherwise it will return -1 and a bool of false.
func Find(slice []string, val []string) ([]string, bool) {
	var dirs []string

	for _, item := range slice {
		if item == val[0] {
			dirs = append(dirs, val[0])
		} else if item == val[1] {
			dirs = append(dirs, val[1])
		}
	}

	if dirs != nil {
		return dirs, true
	}

	return nil, false
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
