// Package sysutil provides utility functions needed for the cloud-platform-applier
package util

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetFolderChunks(repoPath string, numRoutines int) ([][]string, error) {

	folders, err := ListFolderPaths(repoPath)
	if err != nil {
		return nil, err
	}

	nsFolders := []string{}

	// skip the root folder namespaces/cluster.cloud-platform.service.justice.gov.uk which is the first
	// element of the slice. We dont want to apply from the root folder
	for _, f := range folders[1:] {
		nsFolders = append(nsFolders, f)
	}

	folderChunks, err := chunkFolders(nsFolders, numRoutines)
	if err != nil {
		return nil, err
	}
	return folderChunks, nil
}

// ListFolders take the path as input, list all the folders in the give path and
// return a array of strings containing the list of folders
func ListFolderPaths(path string) ([]string, error) {
	var folders []string

	err := filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.Name() == ".terraform" || info.Name() == "resources" {
				return filepath.SkipDir
			}

			if info.IsDir() {
				folders = append(folders, path)
			}
			return nil
		})
	if err != nil {
		return nil, err
	}

	return folders, nil
}

func chunkFolders(folders []string, nRoutines int) ([][]string, error) {

	nChunks := len(folders) / nRoutines

	fmt.Println("Number of folders per chunk", nChunks)

	var folderChunks [][]string
	for {
		if len(folders) == 0 {
			break
		}

		if len(folders) < nChunks {
			nChunks = len(folders)
		}

		folderChunks = append(folderChunks, folders[0:nChunks])
		folders = folders[nChunks:]
	}
	return folderChunks, nil
}

func ListFiles(path string) ([]string, error) {

	var files []string

	err := filepath.WalkDir(path, func(path string, dir os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(dir.Name()) == ".yaml" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil

}
