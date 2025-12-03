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

var (
	defaultTagsPattern = regexp.MustCompile(`\s*default_tags\s*=?\s*{`)
	tagsPattern        = regexp.MustCompile(`\s*tags\s*=\s*{`)
	providerAwsPattern = regexp.MustCompile(`provider\s+"aws"\s*{`)
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
		return fmt.Errorf("no namespaces found for environment: %s", namespacesToProcess)
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
		searchTags: []string{"business-unit", "application", "is-production", "owner", "namespace", "service-area"},
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

	inProviderBlock := false
	inDefaultTagsBlock := false
	providerBraceCount := 0
	defaultTagsBraceCount := 0
	currentProviderHasDefaultTags := false
	tagsAdded := false

	for i, line := range lines {
		// Check if we're entering an AWS provider block
		if providerAwsPattern.MatchString(line) {
			inProviderBlock = true
			providerBraceCount = strings.Count(line, "{") - strings.Count(line, "}")
			currentProviderHasDefaultTags = false
			tagsAdded = false
			newLines = append(newLines, line)
			continue
		}

		// Track brace count in provider block
		if inProviderBlock {
			providerBraceCount += strings.Count(line, "{")
			providerBraceCount -= strings.Count(line, "}")

			// Check if we're entering a default_tags block
			if defaultTagsPattern.MatchString(line) && !inDefaultTagsBlock {
				inDefaultTagsBlock = true
				currentProviderHasDefaultTags = true
				defaultTagsBraceCount = strings.Count(line, "{") - strings.Count(line, "}")
				newLines = append(newLines, line)
				continue
			}

			// Inside default_tags block
			if inDefaultTagsBlock {
				defaultTagsBraceCount += strings.Count(line, "{")
				defaultTagsBraceCount -= strings.Count(line, "}")

				// Check if we're in the tags section
				if tagsPattern.MatchString(line) && !tagsAdded {
					newLines = append(newLines, line)
					// Add missing tags
					for _, tag := range missingTags {
						tagValue := tc.getTagValue(tag)
						tagLine := fmt.Sprintf("      %s = %s", tag, tagValue)
						newLines = append(newLines, tagLine)
					}
					tagsAdded = true
					continue
				}

				newLines = append(newLines, line)

				// Check if we're exiting default_tags block
				if defaultTagsBraceCount == 0 {
					inDefaultTagsBlock = false
				}
				continue
			}

			// If we're closing the provider block and it didn't have default_tags, add it
			if providerBraceCount == 1 && !currentProviderHasDefaultTags && !tagsAdded {
				// Insert default_tags block before the closing brace
				// Check if the next line is the closing brace
				if i+1 < len(lines) && strings.TrimSpace(lines[i+1]) == "}" {
					newLines = append(newLines, "  default_tags {")
					newLines = append(newLines, "    tags = {")
					for _, tag := range tc.searchTags {
						tagValue := tc.getTagValue(tag)
						tagLine := fmt.Sprintf("      %s = %s", tag, tagValue)
						newLines = append(newLines, tagLine)
					}
					newLines = append(newLines, "    }")
					newLines = append(newLines, "  }")
					tagsAdded = true
				}
			}

			newLines = append(newLines, line)

			// Check if we're exiting provider block
			if providerBraceCount == 0 {
				inProviderBlock = false
			}
			continue
		}

		// Not in any special block, just copy the line
		newLines = append(newLines, line)
	}

	newContent := strings.Join(newLines, "\n")
	err := os.WriteFile(filePath, []byte(newContent), 0o644)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
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
	case "service-area":
		return "var.service_area"
	default:
		return fmt.Sprintf("var.%s", strings.ReplaceAll(tag, "-", "_"))
	}
}
