package terraform

import (
	"bufio"
	"os"
	"path/filepath"
)

func find(slice []string, val []string) ([]string, bool) {
	var dirs []string

	for _, item := range slice {
		for _, d := range val {
			if item == d {
				dirs = append(dirs, d)
			}
		}
	}

	if dirs != nil {
		return dirs, true
	}

	return nil, false
}

func targetDirs(file string) ([]string, error) {
	var dirs []string // Directories where tf plan is going to be executed

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if contains(dirs, filepath.Dir(scanner.Text())) != true {
			dirs = append(dirs, filepath.Dir(scanner.Text()))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return dirs, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
