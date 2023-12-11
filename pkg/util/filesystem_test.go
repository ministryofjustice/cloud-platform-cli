package util_test

import (
	"log"
	"os"
	"reflect"
	"strconv"
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

func createTestFolderStructure(t *testing.T) {
	// loop and create folders
	for i := 0; i < 10; i++ {
		tempDir := "namespaces/testCluster/testNamespace" + strconv.Itoa(i)
		tempFile := tempDir + "/foo.yaml"

		if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
			t.Fatal(err)
		}

		_, err := os.Create(tempFile)
		if err != nil {
			t.Fatal(err)
		}
	}

}

func TestGetFolderChunks(t *testing.T) {
	createTestFolderStructure(t)

	defer os.RemoveAll("namespaces")

	type args struct {
		repoPath   string
		batchIndex int
		batchSize  int
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "Get 3 folders starting from index 3",
			args: args{
				repoPath:   "namespaces",
				batchIndex: 3,
				batchSize:  3,
			},
			want: []string{
				"namespaces/testCluster/testNamespace2",
				"namespaces/testCluster/testNamespace3",
				"namespaces/testCluster/testNamespace4",
			},
			wantErr: false,
		},
		{
			name: "Get 3 folders starting from index 9 of total 10 namespaces",
			args: args{
				repoPath:   "namespaces",
				batchIndex: 9,
				batchSize:  3,
			},
			want: []string{
				"namespaces/testCluster/testNamespace8",
				"namespaces/testCluster/testNamespace9",
			},
			wantErr: false,
		},
		{
			name: "Get 3 folders starting from index 20",
			args: args{
				repoPath:   "namespaces",
				batchIndex: 20,
				batchSize:  3,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Get -3 folders starting from index 0",
			args: args{
				repoPath:   "namespaces",
				batchIndex: 0,
				batchSize:  -3,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Get 3 folders starting from index -3",
			args: args{
				repoPath:   "namespaces",
				batchIndex: -3,
				batchSize:  3,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := util.GetFolderChunks(tt.args.repoPath, tt.args.batchIndex, tt.args.batchSize)
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

func TestListFiles(t *testing.T) {
	createTestFolderStructure(t)

	defer os.RemoveAll("namespaces")
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "List files from a folder",
			args: args{
				path: "namespaces/testCluster/testNamespace0",
			},
			want: []string{
				"namespaces/testCluster/testNamespace0/foo.yaml",
			},
			wantErr: false,
		},
		{
			name: "List files from non-existing folder",
			args: args{
				path: "namespaces/testCluster/testNamespace10",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := util.ListFiles(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}
