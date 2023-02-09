package util_test

import (
	"log"
	"os"
	"testing"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/util"
)

func TestListFolders(t *testing.T) {
	repoPath := "somerepo"

	err := os.MkdirAll(repoPath, os.ModePerm)
	if err != nil {
		t.Errorf("Failed to create repo path: %s", err)
	}

	ns_folder := "somerepo/somenamespace"

	err = os.Mkdir(ns_folder, os.ModePerm)
	if err != nil {
		t.Errorf("Failed to create repo path: %s", err)
	}

	folders, err := util.ListFolderPaths(repoPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, folder := range folders {
		t.Logf("Found directory %v\n", folder)
	}

	_, found := util.Find(folders, ns_folder)
	if !found {
		t.Errorf("Expected directory %v not found", ns_folder)
	}
	defer os.RemoveAll(repoPath)
}

func TestIsFilePathExists(t *testing.T) {
	tempDir := "namespaces/testCluster/testNamespace"
	tempFile := tempDir + "/resources"

	if err := os.MkdirAll(tempFile, os.ModePerm); err != nil {
		t.Fatal(err)
	}

	got, err := util.IsFilePathExists(tempDir)
	if err != nil {
		t.Errorf("IsFilePathExists() error = %v", err)
	}
	if !got {
		t.Errorf("Expected file %v not found", tempFile)
	}
	defer os.RemoveAll("namespaces")
}
