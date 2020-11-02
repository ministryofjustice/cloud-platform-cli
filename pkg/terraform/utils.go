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
	eks := []string{"terraform/cloud-platform-eks/components", "terraform/cloud-platform-eks"}

	dirsToPlanKops, foundKops := Find(dirs, kops)
	err := TestPlan(c, "kops", foundKops, dirsToPlanKops)
	if err != nil {
		return err
	}

	dirsToPlanEks, foundEks := Find(dirs, eks)
	err = TestPlan(c, "eks", foundEks, dirsToPlanEks)
	if err != nil {
		return err
	}

	return nil
}

func TestPlan(c *Commander, eks_or_kops string, f bool, d []string) error {
	if c.Environment == eks_or_kops && f {
		for _, dir := range d {
			fmt.Println(dir)
			c.cmdDir = dir
			err := c.Plan()
			if err != nil {
				fmt.Println(err)
				return err
			}
		}
	}
	return nil
}

// Find takes a slice and looks for an element in it. If found it will
// return it's key, otherwise it will return -1 and a bool of false.
func Find(slice []string, val []string) ([]string, bool) {
	var dirs []string

	for _, item := range slice {
		for _, d := range val {
			if item == d {
				dirs = append(dirs, d)
			}
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
