package environment

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

func fileContainsString(t *testing.T, filename string, searchString string) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if !(strings.Contains(string(contents), searchString)) {
		t.Errorf(fmt.Sprintf("Didn't find %s in contents of %s", searchString, filename))
	}
}
