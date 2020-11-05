package terraform

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestTargetDirs confirms given a file directory (string) in a text file,
// the targetDirs function will extract the directory path successfully.
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
