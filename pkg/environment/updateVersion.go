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

func updateVersion(moduleName, actualDbVersion, terraformDbVersion, tfDir string) ([]string, error) {
	log.Printf("Checking if directory exists: %s", tfDir)
	if _, err := os.Stat(tfDir); os.IsNotExist(err) {
		abs, _ := filepath.Abs(tfDir)
		return nil, fmt.Errorf("directory does not exist: %s (resolved to: %s)", tfDir, abs)
	}

	grepCmd := fmt.Sprintf("grep -l 'module \"%s\"' *.tf", moduleName)
	cmd := exec.Command("/bin/sh", "-c", grepCmd)
	cmd.Dir = tfDir
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to find module file: %w", err)
	}
	fileName := strings.TrimSpace(strings.Split(string(output), "\n")[0])

	var changedFiles []string

	log.Printf("Updating hardcoded db_engine_version in file: %s", fileName)
	hardcodeUpdated, err := updateVersionHardcode(fileName, actualDbVersion, terraformDbVersion, tfDir)
	if err != nil {
		return nil, err
	}
	if hardcodeUpdated {
		changedFiles = append(changedFiles, fileName)
	}

	log.Printf("Updating variable db_engine_version in file: %s", fileName)
	varUpdated, varFile, err := updateVersionVariable(fileName, moduleName, actualDbVersion, terraformDbVersion, tfDir)
	if err != nil {
		return nil, err
	}
	if varUpdated && varFile != "" && varFile != fileName {
		changedFiles = append(changedFiles, varFile)
	}

	if len(changedFiles) == 0 {
		return nil, fmt.Errorf("no changes applied for module: %s", moduleName)
	}

	return changedFiles, nil
}

func updateVersionHardcode(fileName, actualDbVersion, terraformDbVersion, tfDir string) (bool, error) {
	sedCmd := fmt.Sprintf("sed -i -e 's/db_engine_version *= *\"%s\"/db_engine_version = \"%s\"/g' %s", terraformDbVersion, actualDbVersion, fileName)
	cmd := exec.Command("/bin/sh", "-c", sedCmd)
	cmd.Dir = tfDir
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("failed to update hardcoded version: %w", err)
	}
	log.Printf("Hardcoded db_engine_version updated successfully in %s", fileName)
	return true, nil
}

func updateVersionVariable(fileName, moduleName, actualDbVersion, terraformDbVersion, tfDir string) (bool, string, error) {
	scanCmd := fmt.Sprintf("grep -A 40 'module \"%s\"' %s", moduleName, fileName)
	cmd := exec.Command("/bin/sh", "-c", scanCmd)
	cmd.Dir = tfDir
	output, err := cmd.Output()
	if err != nil {
		return false, "", fmt.Errorf("failed to scan module block: %w", err)
	}

	re := regexp.MustCompile(`(?i)db_engine_version *= *var\.([a-zA-Z0-9_\-]+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		log.Printf("No variable found for db_engine_version in module %s", moduleName)
		return false, "", nil
	}
	varName := matches[1]

	grepVarCmd := fmt.Sprintf("grep -l 'variable \"%s\"' *.tf", varName)
	varCmd := exec.Command("/bin/sh", "-c", grepVarCmd)
	varCmd.Dir = tfDir
	varOutput, err := varCmd.Output()
	if err != nil {
		return false, "", fmt.Errorf("failed to find variable definition for %s: %w", varName, err)
	}
	varFile := strings.TrimSpace(string(varOutput))

	varSedCmd := fmt.Sprintf("sed -i -e 's/default = \"%s\"/default = \"%s\"/' %s", terraformDbVersion, actualDbVersion, varFile)
	updateVarCmd := exec.Command("/bin/sh", "-c", varSedCmd)
	updateVarCmd.Dir = tfDir
	if err := updateVarCmd.Run(); err != nil {
		return false, "", fmt.Errorf("failed to update variable default: %w", err)
	}
	log.Printf("Variable db_engine_version updated successfully in %s", varFile)
	return true, varFile, nil
}
