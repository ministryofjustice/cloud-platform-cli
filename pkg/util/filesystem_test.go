package util_test

import (
	"log"
	"os"
	"reflect"
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

func TestIsYamlFileExists(t *testing.T) {
	tempDir := "namespaces/testCluster/testNamespace"
	tempFile := tempDir + "/foo.yaml"

	if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
		t.Fatal(err)
	}

	got := util.IsYamlFileExists(tempDir)
	if got {
		t.Errorf("Expected no yaml file %v got %v", tempFile, got)
	}

	_, err := os.Create(tempFile)
	if err != nil {
		t.Fatal(err)
	}

	got = util.IsYamlFileExists(tempDir)
	if !got {
		t.Errorf("Expected file %v not found", tempFile)
	}
	defer os.RemoveAll("namespaces")
}

func TestGetFolderChunks(t *testing.T) {
	tempDir1 := "namespaces/ns1/.terraform"

	if err := os.MkdirAll(tempDir1, os.ModePerm); err != nil {
		t.Fatal(err)
	}
	tempDir2 := "namespaces/ns2/.terraform"

	if err := os.MkdirAll(tempDir2, os.ModePerm); err != nil {
		t.Fatal(err)
	}
	tempDir3 := "namespaces/ns3/.terraform"

	if err := os.MkdirAll(tempDir3, os.ModePerm); err != nil {
		t.Fatal(err)
	}

	type args struct {
		repoPath string
		nsStart  int
		nsEnd    int
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "GetFolderChunks",
			args: args{
				repoPath: "namespaces",
				nsStart:  0,
				nsEnd:    1,
			},
			want:    []string{"namespaces/ns1", "namespaces/ns2"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := util.GetFolderChunks(tt.args.repoPath, tt.args.nsStart, tt.args.nsEnd)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFolderChunks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFolderChunks() = %v, want %v", got, tt.want)
			}
		})
	}
}
