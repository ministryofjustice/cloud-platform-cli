package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v68/github"
)

func (c *GithubClient) SearchAuthTypeDefaultInPR(ctx context.Context, prNumber int) (string, error) {
	opt := &github.ListOptions{PerPage: 100}
	for {
		files, resp, err := c.PullRequests.ListFiles(ctx, c.Owner, c.Repository, prNumber, opt)
		if err != nil {
			return "", fmt.Errorf("error listing files in pull request: %v", err)
		}
		for _, file := range files {
			if file.GetFilename() == "variables.tf" && file.GetPatch() != "" {
				// Try to extract default value for variable "auth_type"
				if defVal := extractAuthTypeDefault(file.GetPatch()); defVal != "" {
					return defVal, nil
				}
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return "", fmt.Errorf("auth_type variable with default not found in PR")
}

// extractAuthTypeDefault parses a patch and returns the default value for variable "auth_type" if present
func extractAuthTypeDefault(patch string) string {
	// Look for lines like: +variable "auth_type" { ... default = "something" ... }
	var inVarBlock bool
	for _, line := range splitLines(patch) {
		if !inVarBlock && (line == "+variable \"auth_type\" {" || line == "+variable \"auth_type\"{") {
			inVarBlock = true
		} else if inVarBlock {
			if line == "+}" {
				inVarBlock = false
			} else if len(line) > 0 && (line[0] == '+' || line[0] == ' ') {
				// Look for default assignment
				if _, val, ok := parseDefaultLine(line); ok {
					return val
				}
			}
		}
	}
	return ""
}

// parseDefaultLine tries to parse a line like '+  default = "something"' and returns the value
func parseDefaultLine(line string) (string, string, bool) {
	// Remove leading + and whitespace
	l := line
	if len(l) > 0 && l[0] == '+' {
		l = l[1:]
	}
	l = trimSpace(l)
	if len(l) >= 8 && l[:8] == "default=" {
		l = l[8:]
	} else if len(l) >= 9 && l[:8] == "default " {
		l = l[8:]
		if l[0] == '=' {
			l = l[1:]
		}
	} else if len(l) >= 10 && l[:9] == "default =" {
		l = l[9:]
	} else {
		return "", "", false
	}
	l = trimSpace(l)
	// Remove quotes if present
	if len(l) > 1 && l[0] == '"' && l[len(l)-1] == '"' {
		return "default", l[1 : len(l)-1], true
	}
	return "default", l, true
}

// splitLines splits a string into lines
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// trimSpace trims leading and trailing whitespace
func trimSpace(s string) string {
	i, j := 0, len(s)-1
	for i <= j && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	for j >= i && (s[j] == ' ' || s[j] == '\t') {
		j--
	}
	return s[i : j+1]
}
