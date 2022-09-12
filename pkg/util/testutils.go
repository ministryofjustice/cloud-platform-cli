package util

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func FileContainsString(t *testing.T, filename string, searchString string) {
	contents, err := os.ReadFile(filename)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if !(strings.Contains(string(contents), searchString)) {
		t.Errorf(fmt.Sprintf("Didn't find string: %s in file: %s", searchString, filename))
	}
}

func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}
