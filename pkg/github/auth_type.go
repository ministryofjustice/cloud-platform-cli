package github

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v74/github"
)

func (c *GithubClient) FlagCheckAuthType(ctx context.Context, prNumber int, namespace, clusterName string) (string, error) {
	var branch string
	if prNumber == 0 && namespace == "" {
		return "", fmt.Errorf("either -pr-number or -namespace flag is required")
	} else if prNumber > 0 {
		// get namespace from PR
		prDetails, err := c.PRDetails(context.Background(), prNumber, clusterName)
		if err != nil {
			return "", fmt.Errorf("failed to get pr details: %v", err)
		}
		branch = prDetails[0]
		namespace = prDetails[1]
		if namespace == "" {
			return "", fmt.Errorf("namespace not found in pr %d", prNumber)
		}
	} else if namespace != "" {
		// get branch from from local
		branch = getCurrentBranch()
	}

	// get authtype this is only needed for migration purposes once users are all using github app this can be removed
	authType, err := c.SearchAuthTypeInRepo(context.Background(), namespace, branch, clusterName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get auth_type from PR: %v, defaulting to token auth\n", err) // DEBUG
		authType = "token"
	}

	return authType, nil
}

func getCurrentBranch() string {
	branchBytes, err := os.ReadFile(".git/HEAD")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read .git/HEAD: %v, defaulting to main\n", err) // DEBUG
		return "main"
	}
	branchRef := strings.TrimSpace(string(branchBytes))
	if strings.HasPrefix(branchRef, "ref: refs/heads/") {
		return strings.TrimPrefix(branchRef, "ref: refs/heads/")
	}
	fmt.Fprintf(os.Stderr, "Unexpected format in .git/HEAD: %s, defaulting to main\n", branchRef) // DEBUG
	return "main"
}

func (c *GithubClient) PRDetails(ctx context.Context, prNumber int, clusterName string) ([]string, error) {
	var namespace string
	pr, _, err := c.PullRequests.Get(ctx, c.Owner, c.Repository, prNumber)
	if err != nil {
		return nil, fmt.Errorf("error getting PR %d: %v", prNumber, err)
	}
	branch := pr.GetHead().GetRef()
	files, _, err := c.PullRequests.ListFiles(ctx, c.Owner, c.Repository, prNumber, nil)
	if err != nil {
		return nil, fmt.Errorf("error listing files for PR %d: %v", prNumber, err)
	}
	for _, file := range files {
		// split file path by "/"
		pathParts := strings.Split(file.GetFilename(), "/")
		// check if path matches expected pattern
		if len(pathParts) >= 5 && pathParts[0] == "namespaces" && pathParts[1] == clusterName+".cloud-platform.service.justice.gov.uk" {
			namespace = pathParts[2]
			break
		}
	}

	prDetails := []string{branch, namespace}
	return prDetails, nil
}

// search repo for auth_type variable default in a PR depending on namespace directory name
func (c *GithubClient) SearchAuthTypeInRepo(ctx context.Context, namespace, branch, clusterName string) (string, error) {
	path := fmt.Sprintf("namespaces/%s.cloud-platform.service.justice.gov.uk/%s/resources/variables.tf", clusterName, namespace)

	// Clean up branch reference - remove head/ prefix if present
	cleanBranch := strings.TrimPrefix(branch, "head/")

	opt := &github.RepositoryContentGetOptions{
		Ref: cleanBranch,
	}

	fileContent, _, _, err := c.V3.Repositories.GetContents(ctx, c.Owner, c.Repository, path, opt)
	if err != nil {
		// Try fallback: check if the file exists in main branch
		mainOpt := &github.RepositoryContentGetOptions{
			Ref: "main",
		}
		var fallbackErr error
		fileContent, _, _, fallbackErr = c.V3.Repositories.GetContents(ctx, c.Owner, c.Repository, path, mainOpt)
		if fallbackErr != nil {
			return "", fmt.Errorf("error getting file %s (tried branch %s and main): %v", path, cleanBranch, err)
		}
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
		if !inVarBlock && len(l) > 0 && l[0] == '+' && containsVarAuthType(l) {
			inVarBlock = true
			continue
		}
		if inVarBlock {
			if l == "+}" || l == "+ }" {
				inVarBlock = false
				continue
			}
			if len(l) > 0 && (l[0] == '+' || l[0] == ' ') {
				_, val, ok := parseDefaultLine(l)
				if ok {
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
			if rest2 == "{" {
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
