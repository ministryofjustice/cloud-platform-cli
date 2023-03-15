package terraform

import (
	"os"
	"testing"
)

func Test_DeleteLocalState(t *testing.T) {
	parentDir := "testParent"
	file := "testFile"
	siblingDir := "testDir"

	os.RemoveAll(parentDir)
	err := os.Mkdir(parentDir, 0755)
	if err != nil {
		t.Errorf("DeleteLocalState() error = %v", err)
	}
	defer os.RemoveAll(parentDir)

	// create file in temp directory
	_, err = os.CreateTemp(parentDir, file)
	if err != nil {
		t.Errorf("DeleteLocalState() error = %v", err)
	}

	// create directory in temp directory
	_, err = os.MkdirTemp(parentDir, siblingDir)
	if err != nil {
		t.Errorf("DeleteLocalState() error = %v", err)
	}

	if err := DeleteLocalState(parentDir, file, siblingDir); err != nil {
		t.Errorf("DeleteLocalState() error = %v", err)
	}

	if _, err = os.Stat(file); !os.IsNotExist(err) {
		t.Errorf("DeleteLocalState() error = %v", "file not deleted")
	}

	if _, err := os.Stat(siblingDir); !os.IsNotExist(err) {
		t.Errorf("DeleteLocalState() error = %v", "directory not deleted")
	}
}
