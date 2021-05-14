package terraform

import (
	"bufio"
	"os"
	"path/filepath"
)

// targetDirs return the directories where terraform plan is going to be executed
func targetDirs(file string) ([]string, error) {
	var dirs []string // Directories where tf plan is going to be executed

	dirsAllowed := []string{
		"terraform/aws-accounts/cloud-platform-aws/vpc/kops/components",
		"terraform/aws-accounts/cloud-platform-aws/vpc/kops",
		"terraform/global-resources",
		"terraform/aws-accounts/cloud-platform-aws/vpc/eks/components",
		"terraform/aws-accounts/cloud-platform-aws/vpc/eks",
		"terraform/cloud-platform-account",
		"terraform/cloud-platform-network",
	}

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// The first condition evaluates if the element already exists in the slice (why to execute
		// plan twice against the same dir?). The second condition evaluates if the element is in
		// the desired list to execute Plan (we don't want to execute Plan against everything)
		if contains(dirs, filepath.Dir(scanner.Text())) != true &&
			contains(dirsAllowed, filepath.Dir(scanner.Text())) == true {
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
