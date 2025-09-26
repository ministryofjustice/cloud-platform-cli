package github

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v74/github"
)

func (c *GithubClient) FlagCheckAuthType(ctx context.Context, prNumber int, namespace string) (string, error) {
	var branch string
	if prNumber == 0 && namespace == "" {
		return "", fmt.Errorf("either -pr-number or -namespace flag is required")
	} else if prNumber > 0 {
		// get namespace from PR
		prDetails, err := c.PRDetails(context.Background(), prNumber)
		if err != nil {
			return "", fmt.Errorf("failed to get pr details: %v", err)
		}
		branch = prDetails[0]
		namespace = prDetails[1]
		if namespace == "" {
			return "", fmt.Errorf("namespace not found in pr %d", prNumber)
		}
	} else if namespace != "" {
		branch = "main"
	}

	// get authtype this is only needed for migration purposes once users are all using github app this can be removed
	authType, err := c.SearchAuthTypeInRepo(context.Background(), namespace, branch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get auth_type from PR: %v, defaulting to token auth\n", err) // DEBUG
		authType = "token"
	}

	return authType, nil
}

func (c *GithubClient) PRDetails(ctx context.Context, prNumber int) ([]string, error) {
	var namespace string
	pr, _, err := c.V3.PullRequests.Get(ctx, c.Owner, c.Repository, prNumber)
	if err != nil {
		return nil, fmt.Errorf("error getting PR %d: %v", prNumber, err)
	}
	branch := pr.GetHead().GetRef()
	fmt.Println(branch)
	files, _, err := c.V3.PullRequests.ListFiles(ctx, c.Owner, c.Repository, prNumber, nil)
	if err != nil {
		return nil, fmt.Errorf("error listing files for PR %d: %v", prNumber, err)
	}
	for _, file := range files {
		fmt.Println(file.Filename)
		// split file path by "/"
		pathParts := strings.Split(file.GetFilename(), "/")
		// check if path matches expected pattern
		if len(pathParts) >= 5 && pathParts[0] == "namespaces" && pathParts[1] == "live.cloud-platform.service.justice.gov.uk" {
			namespace = pathParts[2]
			break
		}
		fmt.Println(namespace)
	}

	prDetails := []string{branch, namespace}
	return prDetails, nil
}

// search repo for auth_type variable default in a PR depending on namespace directory name
func (c *GithubClient) SearchAuthTypeInRepo(ctx context.Context, namespace, branch string) (string, error) {
	path := fmt.Sprintf("namespaces/live.cloud-platform.service.justice.gov.uk/%s/resources/variables.tf", namespace)
	opt := &github.RepositoryContentGetOptions{
		Ref: branch,
	}
	fileContent, _, _, err := c.V3.Repositories.GetContents(ctx, c.Owner, c.Repository, path, opt)
	if err != nil {
		return "", fmt.Errorf("error getting directory contents for %s: %v", path, err)
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return "", fmt.Errorf("error getting file content for %s: %v", path, err)
	}

	// Try to extract default value for variable "auth_type"
	if defVal := extractAuthTypeDefault(content); defVal != "" {
		return defVal, nil
	}

	return "", fmt.Errorf("auth_type variable with default not found in %s", path)
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
