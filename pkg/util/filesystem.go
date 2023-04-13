// Package sysutil provides utility functions needed for the cloud-platform-applier
package util

import (
	"errors"
	"log"
	"os"
	"path/filepath"
)

// GetFolderChunks get the list of folders from the given path, get list of folders from nsStart till nsEnd
func GetFolderChunks(repoPath string, nsStart int, nsEnd int) ([]string, error) {
	folders, err := ListFolderPaths(repoPath)
	if err != nil {
		return nil, err
	}

	// skip the root folder namespaces/cluster.cloud-platform.service.justice.gov.uk which is the first
	// element of the slice. We dont want to apply from the root folder
	var nsFolders []string
	nsFolders = append(nsFolders, folders[1:]...)
	var folderChunks []string

	for i := nsStart; i <= nsEnd; i++ {
		folderChunks = append(folderChunks, nsFolders[i])
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

func IsFilePathExists(filePath string) (bool, error) {
	if _, err := os.Stat(filePath); err == nil {
		return true, nil

	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil

	} else {
		return false, err
	}
}

func IsYamlFileExists(path string) bool {
	files, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		fileExtension := filepath.Ext(path + "/" + file.Name())
		if fileExtension == ".yaml" {
			return true
		}
	}
	return false
}
