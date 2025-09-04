package github

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v74/github"
)

func (c *GithubClient) SearchAuthTypeDefaultInPR(ctx context.Context, prNumber int) (string, error) {
	opt := &github.ListOptions{PerPage: 100}
	for {
		files, resp, err := c.PullRequests.ListFiles(ctx, c.Owner, c.Repository, prNumber, opt)
		if err != nil {
			return "", fmt.Errorf("error listing files in pull request: %v", err)
		}
		for _, file := range files {
			fmt.Fprintf(os.Stderr, "PR FILE: %s patch-len=%d\n", file.GetFilename(), len(file.GetPatch())) // DEBUG
			if strings.HasSuffix(file.GetFilename(), "variables.tf") && file.GetPatch() != "" {
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
	// More robust: allow for extra whitespace and ignore indentation
	var inVarBlock bool
	for _, line := range splitLines(patch) {
		l := trimSpace(line)
		fmt.Fprintf(os.Stderr, "PATCH LINE: %s\n", l) // DEBUG
		if !inVarBlock && len(l) > 0 && l[0] == '+' && containsVarAuthType(l) {
			fmt.Fprintf(os.Stderr, "-- Entering auth_type variable block\n") // DEBUG
			inVarBlock = true
			continue
		}
		if inVarBlock {
			if l == "+}" || l == "+ }" {
				fmt.Fprintf(os.Stderr, "-- Exiting variable block\n") // DEBUG
				inVarBlock = false
				continue
			}
			if len(l) > 0 && (l[0] == '+' || l[0] == ' ') {
				key, val, ok := parseDefaultLine(l)
				if ok {
					fmt.Fprintf(os.Stderr, "-- Found default line: key=%s val=%s\n", key, val) // DEBUG
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
	// Accept various spacings: default=, default =, default   =
	if len(l) >= 7 && l[:7] == "default" {
		rest := l[7:]
		rest = trimSpace(rest)
		if len(rest) > 0 && rest[0] == '=' {
			rest = rest[1:]
			rest = trimSpace(rest)
			l = rest
			fmt.Fprintf(os.Stderr, "parseDefaultLine: found default assignment, value: %s\n", l) // DEBUG
		} else {
			return "", "", false
		}
	} else {
		return "", "", false
	}
	// Remove quotes if present
	if len(l) > 1 && l[0] == '"' && l[len(l)-1] == '"' {
		return "default", l[1 : len(l)-1], true
	}
	return "default", l, true
}

// containsVarAuthType checks if a line contains the start of the auth_type variable block
func containsVarAuthType(line string) bool {
	// Accept: +variable "auth_type" {, +variable "auth_type"{, with any whitespace
	if len(line) < 18 {
		return false
	}
	// Remove leading + and whitespace
	l := line
	if l[0] == '+' {
		l = l[1:]
	}
	l = trimSpace(l)
	if len(l) >= 18 && l[:8] == "variable" {
		rest := l[8:]
		rest = trimSpace(rest)
		if len(rest) >= 12 && rest[:11] == "\"auth_type\"" {
			rest2 := rest[11:]
			rest2 = trimSpace(rest2)
			if rest2 == "{" || rest2 == "{" {
				return true
			}
		}
	}
	return false
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
