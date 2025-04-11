package environment

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// updateVersion updates a Terraform module file with either a hardcoded value change
// or via a variable default update. It returns the list of files that were changed.
func updateVersion(moduleName, actualDbVersion, terraformDbVersion, tfDir string) ([]string, error) {
	if _, err := os.Stat(tfDir); os.IsNotExist(err) {
		abs, _ := filepath.Abs(tfDir)
		return nil, fmt.Errorf("directory does not exist: %s (resolved to: %s)", tfDir, abs)
	}

	// Locate the Terraform file(s) that contain the module.
	grepCmd := fmt.Sprintf("grep -l 'module \"%s\"' *.tf", moduleName)
	cmd := exec.Command("/bin/sh", "-c", grepCmd)
	cmd.Dir = tfDir
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to find module file: %w", err)
	}

	// Log all files returned by grep
	filesFound := strings.Split(strings.TrimSpace(string(output)), "\n")
	log.Printf("Files found for module %q: %v", moduleName, filesFound)

	// For this example, we work with the first file in the list.
	if len(filesFound) == 0 || filesFound[0] == "" {
		return nil, fmt.Errorf("module %s not found in any .tf file", moduleName)
	}
	fileName := strings.TrimSpace(filesFound[0])
	log.Printf("Processing module %q from file: %s", moduleName, fileName)

	// Extract a block of lines that likely contains the module definition.
	scanCmd := fmt.Sprintf("grep -A 40 'module \"%s\"' %s", moduleName, fileName)
	cmd = exec.Command("/bin/sh", "-c", scanCmd)
	cmd.Dir = tfDir
	blockBytes, err := cmd.Output()
	if err != nil {
		log.Printf("Failed to extract module block for %s: %v", moduleName, err)
		return nil, err
	}
	moduleBlock := string(blockBytes)
	log.Printf("Module block for %q:\n%s", moduleName, moduleBlock)

	var changedFiles []string

	// Try to update a hardcoded value first.
	hardcodeUpdated, err := updateVersionHardcode(moduleName, fileName, actualDbVersion, terraformDbVersion, tfDir, moduleBlock)
	if err == nil && hardcodeUpdated {
		changedFiles = append(changedFiles, fileName)
		return changedFiles, nil
	}

	// If the hardcoded update wasn't applied, try to update via variable default.
	varUpdated, varFile, err := updateVersionVariable(fileName, moduleName, actualDbVersion, terraformDbVersion, tfDir, moduleBlock)
	if err != nil {
		return nil, err
	}
	if varUpdated && varFile != "" {
		changedFiles = append(changedFiles, varFile)
		return changedFiles, nil
	}

	return nil, fmt.Errorf("no changes applied for module: %s", moduleName)
}

// updateVersionHardcode updates the hardcoded "db_engine_version" directly inside the module block.
func updateVersionHardcode(moduleName, fileName, actualDbVersion, terraformDbVersion, tfDir string, moduleBlock string) (bool, error) {
	pattern := fmt.Sprintf(`db_engine_version\s*=\s*"%s"`, regexp.QuoteMeta(terraformDbVersion))
	found, err := regexp.MatchString(pattern, moduleBlock)
	if err != nil {
		return false, fmt.Errorf("regex match error: %w", err)
	}
	if found {
		sedCmd := fmt.Sprintf("sed -i -E '/module \"%s\"/,/^}/s/db_engine_version[[:space:]]*=[[:space:]]*\"%s\"/db_engine_version = \"%s\"/' %s",
			moduleName, terraformDbVersion, actualDbVersion, fileName)
		execCmd := exec.Command("/bin/sh", "-c", sedCmd)
		execCmd.Dir = tfDir
		if err := execCmd.Run(); err != nil {
			log.Printf("Hardcoded update failed inside block: %v", err)
			return false, err
		}
		log.Printf("Hardcoded db_engine_version updated in module %q in file %s", moduleName, fileName)
		return true, nil
	}
	return false, nil
}

// updateVersionVariable updates the default value of the variable referenced by db_engine_version.
func updateVersionVariable(fileName, moduleName, actualDbVersion, terraformDbVersion, tfDir string, moduleBlock string) (bool, string, error) {
	// Use a regex that tolerates extra whitespace around '=' to capture the variable name.
	re := regexp.MustCompile(`(?i)db_engine_version\s*=\s*var\.([a-zA-Z0-9_\-]+)`)
	matches := re.FindStringSubmatch(moduleBlock)
	log.Printf("Regex matches for variable extraction: %v", matches)
	if len(matches) < 2 {
		log.Printf("No variable found for db_engine_version in module %q", moduleName)
		return false, "", nil
	}
	// matches[0] is the entire match; matches[1] is the first captured group (the variable name).
	varName := matches[1]
	log.Printf("Detected variable name: %s", varName)

	// Locate the file containing the variable definition.
	grepVarCmd := fmt.Sprintf("grep -l 'variable \"%s\"' *.tf", varName)
	varCmd := exec.Command("/bin/sh", "-c", grepVarCmd)
	varCmd.Dir = tfDir
	varOutput, err := varCmd.Output()
	if err != nil {
		return false, "", fmt.Errorf("failed to find variable definition for %s: %w", varName, err)
	}
	varFile := strings.TrimSpace(string(varOutput))
	log.Printf("Found variable %q in file: %s", varName, varFile)

	// Use sed with extended regex to update the default value in the variable definition.
	varSedCmd := fmt.Sprintf("sed -i -E 's/default[[:space:]]*=[[:space:]]*\"%s\"/default = \"%s\"/' %s",
		terraformDbVersion, actualDbVersion, varFile)
	updateVarCmd := exec.Command("/bin/sh", "-c", varSedCmd)
	updateVarCmd.Dir = tfDir
	if err := updateVarCmd.Run(); err != nil {
		return false, "", fmt.Errorf("failed to update variable default: %w", err)
	}

	// After updating, print out the updated variable definition.
	grepUpdatedCmd := fmt.Sprintf("grep -A 2 'variable \"%s\"' %s", varName, varFile)
	getVarOutputCmd := exec.Command("/bin/sh", "-c", grepUpdatedCmd)
	getVarOutputCmd.Dir = tfDir
	updatedOutput, err := getVarOutputCmd.Output()
	if err != nil {
		log.Printf("Failed to print updated variable definition: %v", err)
	} else {
		fmt.Printf("Updated variable definition for %s:\n%s\n", varName, string(updatedOutput))
	}

	log.Printf("Variable %q updated successfully in %s", varName, varFile)
	return true, varFile, nil
}
