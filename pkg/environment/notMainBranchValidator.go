package environment

import (
	"fmt"
	"os/exec"
	"strings"
)

type notMainBranchValidator struct{}

func (v *notMainBranchValidator) isValid(s string) bool {
	if s == "main" {
		fmt.Println("Branch cannot be main")
		return false
	}
	if s == "" {
		fmt.Println("A Value is required")
		return false
	}

	branch, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		fmt.Println("Cannot get the git branch")
		return false
	}
	str := strings.Trim(string(branch), "\n")
	if str != s {
		fmt.Println("Current working branch should be the same as the branch name provided")
		return false
	}
	return true
}
