package environment

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type TagChecker struct {
	searchTags []string
	baseDir    string
}

func NamespaceTagging(opt Options) error {
	baseDir := filepath.Clean(opt.RepoPath)
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		log.Fatalf("The specified repository path does not exist: %s", baseDir)
	}

	tagChecker := newTagChecker(baseDir)

	var namespacesToProcess []string
	requestedNamespaces := strings.Split(strings.Join(opt.Namespaces, ","), ",")
	for _, ns := range requestedNamespaces {
		ns = strings.TrimSpace(ns)
		if ns != "" {
			for _, existingNs := range requestedNamespaces {
				if existingNs == ns {
					namespacesToProcess = append(namespacesToProcess, ns)
					break
				}
			}
		}
	}
	fmt.Printf("\nProcessing %d namespaces: %v\n", len(namespacesToProcess), namespacesToProcess)

	if len(namespacesToProcess) == 0 {
		return fmt.Errorf("No namespaces found for environment: %s\n", namespacesToProcess)

	}

	fmt.Printf("\nDo you want to check and add missing tags for %s environment(s)? (y/N): ", namespacesToProcess)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	response := strings.ToLower(strings.TrimSpace(scanner.Text()))

	if response != "y" && response != "yes" {
		fmt.Printf("Exiting without making changes.")
		return nil
	}

	fmt.Printf("\nChecking tags for %s namespaces...\n", namespacesToProcess)

	for i, ns := range namespacesToProcess {
		fmt.Printf("[%d/%d] Processing namespace: %s\n", i+1, len(namespacesToProcess), ns)
		if err := tagChecker.checkAndAddTags(ns); err != nil {
			fmt.Printf("Error processing namespace %s: %v\n", ns, err)
		}
	}
	return nil
}

func newTagChecker(baseDir string) *TagChecker {
	return &TagChecker{
		searchTags: []string{"business-unit", "application", "is-production", "owner", "namespace"},
		baseDir:    baseDir,
	}
}

func (tc *TagChecker) checkAndAddTags(namespace string) error {
	// Find the terraform file path
	tfFilePath, err := tc.findTerraformFile(namespace)
	if err != nil {
		fmt.Printf("Terraform file not found for namespace: %s (%v)\n", namespace, err)
		return nil // Don't treat this as a fatal error
	}

	// Check if file exists
	if _, err := os.Stat(tfFilePath); os.IsNotExist(err) {
		fmt.Printf("Terraform file not found for namespace: %s\n", namespace)
		return nil
	}

	// Read the file
	content, err := os.ReadFile(tfFilePath)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", tfFilePath, err)
	}

	// Check for missing tags
	missingTags := tc.findMissingTags(string(content))

	if len(missingTags) > 0 {
		fmt.Printf("Namespace: %s is missing tags: %v\n", namespace, missingTags)

		// Add missing tags
		err = tc.addMissingTags(tfFilePath, string(content), missingTags)
		if err != nil {
			return fmt.Errorf("error adding tags to %s: %w", tfFilePath, err)
		}

		for _, tag := range missingTags {
			fmt.Printf("Adding tag: %s to %s\n", tag, tfFilePath)
		}
	} else {
		fmt.Printf("Namespace: %s has all default tags.\n", namespace)
	}

	return nil
}

func (tc *TagChecker) findTerraformFile(namespace string) (string, error) {
	// Try multiple glob patterns to find the terraform file
	patterns := []string{
		filepath.Join(tc.baseDir, "namespaces", "live.*", "*", namespace, "resources", "main.tf"),
		filepath.Join(tc.baseDir, "namespaces", "live-2.*", "*", namespace, "resources", "main.tf"),
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}

		if len(matches) > 0 {
			return matches[0], nil
		}
	}

	// If not found, try a more comprehensive search
	return tc.findTerraformFileRecursive(namespace)
}

func (tc *TagChecker) findTerraformFileRecursive(namespace string) (string, error) {
	var foundPath string

	namespacesDir := filepath.Join(tc.baseDir, "namespaces")

	err := filepath.Walk(namespacesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking despite errors
		}

		// Check if this is the main.tf file we're looking for
		if info.Name() == "main.tf" {
			// Check if this path contains our namespace and resources directory
			if strings.Contains(path, namespace) && strings.Contains(path, "resources") {
				// Verify this is actually the right namespace directory structure
				dir := filepath.Dir(path)
				if filepath.Base(dir) == "resources" {
					parentDir := filepath.Dir(dir)
					if filepath.Base(parentDir) == namespace {
						foundPath = path
						return filepath.SkipDir // Found it, stop walking this branch
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	if foundPath == "" {
		return "", fmt.Errorf("no terraform file found for namespace %s", namespace)
	}

	return foundPath, nil
}

func (tc *TagChecker) findMissingTags(content string) []string {
	var missingTags []string

	// Extract default_tags blocks from providers
	defaultTagsPattern := `(?s)default_tags\s*\{[^}]*tags\s*=\s*\{[^}]*\}`
	defaultTagsRegex := regexp.MustCompile(defaultTagsPattern)
	defaultTagsMatches := defaultTagsRegex.FindAllString(content, -1)

	// Also check for default_tags = { ... } format in providers
	defaultTagsPatternAlt := `(?s)default_tags\s*=\s*\{[^}]*\}`
	defaultTagsRegexAlt := regexp.MustCompile(defaultTagsPatternAlt)
	defaultTagsMatchesAlt := defaultTagsRegexAlt.FindAllString(content, -1)

	// Extract default_tags from locals block (more common pattern)
	localsDefaultTagsPattern := `(?s)locals\s*\{[^}]*default_tags\s*=\s*\{[^}]*\}`
	localsDefaultTagsRegex := regexp.MustCompile(localsDefaultTagsPattern)
	localsDefaultTagsMatches := localsDefaultTagsRegex.FindAllString(content, -1)

	// Combine all matches
	allDefaultTagsBlocks := append(defaultTagsMatches, defaultTagsMatchesAlt...)
	allDefaultTagsBlocks = append(allDefaultTagsBlocks, localsDefaultTagsMatches...)

	// If no default_tags blocks found, all tags are missing
	if len(allDefaultTagsBlocks) == 0 {
		return tc.searchTags
	}

	// Combine all default_tags content
	combinedTagsContent := strings.Join(allDefaultTagsBlocks, " ")

	for _, tag := range tc.searchTags {
		// Check for the tag in various formats within the combined tags content
		patterns := []string{
			fmt.Sprintf(`\b%s\s*=`, tag), // business-unit =
			fmt.Sprintf(`"%s"\s*=`, tag), // "business-unit" =
			fmt.Sprintf(`'%s'\s*=`, tag), // 'business-unit' =
		}

		found := false
		for _, pattern := range patterns {
			if matched, _ := regexp.MatchString(pattern, combinedTagsContent); matched {
				found = true
				break
			}
		}

		if !found {
			missingTags = append(missingTags, tag)
		}
	}

	return missingTags
}

func (tc *TagChecker) addMissingTags(filePath, content string, missingTags []string) error {
	lines := strings.Split(content, "\n")
	var newLines []string

	defaultTagsFound := false
	inDefaultTagsBlock := false
	braceCount := 0
	tagsSectionFound := false

	for _, line := range lines {
		newLines = append(newLines, line)

		if matched, _ := regexp.MatchString(`\s*default_tags\s*=?\s*{`, line); matched {
			defaultTagsFound = true
			inDefaultTagsBlock = true
			braceCount = 1
		} else if inDefaultTagsBlock {
			braceCount += strings.Count(line, "{")
			braceCount -= strings.Count(line, "}")
			if matched, _ := regexp.MatchString(`\s*tags\s*=\s*{`, line); matched {
				tagsSectionFound = true
				for _, tag := range missingTags {
					tagValue := tc.getTagValue(tag)
					tagLine := fmt.Sprintf("      %s = %s", tag, tagValue)
					newLines = append(newLines, tagLine)
				}
			}

			if braceCount == 0 {
				inDefaultTagsBlock = false
			}
		}
	}

	if !defaultTagsFound {
		newLines = tc.addDefaultTagsBlock(lines)
	} else if !tagsSectionFound {
		fmt.Printf("Warning: default_tags block found but no tags section in %s\n", filePath)
	}

	newContent := strings.Join(newLines, "\n")
	err := os.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

func (tc *TagChecker) addDefaultTagsBlock(lines []string) []string {
	var newLines []string
	providerBlockFound := false

	for _, line := range lines {
		newLines = append(newLines, line)

		if matched, _ := regexp.MatchString(`provider\s+"aws"\s*{`, line); matched {
			providerBlockFound = true

			newLines = append(newLines, "  default_tags {")
			newLines = append(newLines, "    tags = {")

			for _, tag := range tc.searchTags {
				tagValue := tc.getTagValue(tag)
				tagLine := fmt.Sprintf("      %s = %s", tag, tagValue)
				newLines = append(newLines, tagLine)
			}

			newLines = append(newLines, "    }")
			newLines = append(newLines, "  }")
		}
	}

	if !providerBlockFound {
		fmt.Printf("Warning: No provider 'aws' block found, manual intervention may be required\n")
	}

	return newLines
}

func (tc *TagChecker) getTagValue(tag string) string {
	switch tag {
	case "business-unit":
		return "var.business_unit"
	case "is-production":
		return "var.is_production"
	case "owner":
		return "var.team_name"
	case "application":
		return "var.application"
	case "namespace":
		return "var.namespace"
	default:
		return fmt.Sprintf("var.%s", strings.ReplaceAll(tag, "-", "_"))
	}
}
