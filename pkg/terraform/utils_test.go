package terraform

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestFind tests an element within slice of string is found in another slice of string
func TestFind(t *testing.T) {
	dirsWithChanges := []string{
		"terraform/cloud-platform",
		"terraform/cloud-platform-components",
		"smoke-tests/spec",
		"terraform/cloud-platform-eks/components/",
		"terraform/cloud-platform-eks",
	}

	targetDir := []string{
		"terraform/cloud-platform-eks",
	}

	_, found := find(dirsWithChanges, targetDir)

	if found != true {
		t.Errorf("find() failed. Couldn't find desided directories inside list")
	}

	targetDir = []string{
		"dir/that/doesnt/exist",
	}

	_, found = find(dirsWithChanges, targetDir)

	if found {
		t.Errorf("find() failed. Couldn't find desided directories inside list")
	}
}

// TestTargetDirs tests that confirms given a file directory
// in a text file, the targetDirs function will extract the directory path successfully.
func TestTargetDirs(t *testing.T) {
	fileName := "changedFiles"
	fileString := "/this/is/a/test/dir"

	// A temp file is created with a string inside
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println(err)
	}
	_, err = file.WriteString(fileString)
	if err != nil {
		fmt.Println(err)
	}
	err = file.Close()
	if err != nil {
		fmt.Println(err)
	}

	// The temp file is passed to the targetDirs function so it can extract the string containing
	// a directory path.
	target, err := targetDirs(fileName)
	if err != nil {
		fmt.Println(err)
	}

	// The value extracted from the text file is compared to its expected value. If false, the
	// test will fail.
	for _, v := range target {
		if strings.Contains(fileString, v) {
			fmt.Println("Test passes")
		} else {
			t.Error("Files do not match, test fails")
		}
	}
	os.Remove(fileName)
}
