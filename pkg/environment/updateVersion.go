package environment

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
)

func updateVersion(moduleName, actualDbVersion, terraformDbVersion, tfDir string) ([]string, error) {
	log.Printf("Inputs - moduleName: %s, actualDbVersion: '%s', terraformDbVersion: '%s', tfDir: %s", moduleName, actualDbVersion, terraformDbVersion, tfDir)

	grepCmd := fmt.Sprintf("grep -l 'module \"%s\"' *.tf", moduleName)
	cmd := exec.Command("/bin/sh", "-c", grepCmd)
	cmd.Dir = tfDir
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to find module file: %w", err)
	}
	fileName := strings.TrimSpace(string(output))
	log.Printf("Found module file: %s", fileName)

	scanCmd := fmt.Sprintf("grep -A 40 'module \"%s\"' %s", moduleName, fileName)
	log.Printf("Scanning module block: %s", scanCmd)
	cmd = exec.Command("/bin/sh", "-c", scanCmd)
	cmd.Dir = tfDir
	blockBytes, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	moduleBlock := string(blockBytes)

	lineRegex := regexp.MustCompile(`(?i)db_engine_version\s*=\s*(.+)`)
	lineMatches := lineRegex.FindStringSubmatch(moduleBlock)

	dbValue := strings.TrimSpace(lineMatches[1])
	log.Printf("Found db_engine_version value: %s", dbValue)

	var changedFiles []string

	if strings.Contains(dbValue, "var.") {
		varName := strings.TrimSpace(strings.TrimPrefix(dbValue, "var."))
		updated, varFile, err := updateVersionVariable(actualDbVersion, terraformDbVersion, tfDir, varName)
		if err != nil {
			log.Printf("[ERROR] Failed to update variable version: %v", err)
			return nil, err
		}
		if updated && varFile != "" {
			log.Printf("Updated variable in file: %s", varFile)
			changedFiles = append(changedFiles, varFile)
		}
	} else {
		updated, err := updateVersionHardcode(moduleName, fileName, actualDbVersion, terraformDbVersion, tfDir)
		if err != nil {
			log.Printf("[ERROR] Failed to update hardcoded version: %v", err)
			return nil, err
		}
		if updated {
			log.Printf("Updated hardcoded version in file: %s", fileName)
			changedFiles = append(changedFiles, fileName)
		}
	}

	return changedFiles, nil
}

func updateVersionHardcode(moduleName, fileName, actualDbVersion, terraformDbVersion, tfDir string) (bool, error) {
	sedCmd := fmt.Sprintf(
		`sed -i -E '/module "%s"/,/^}/ s/"%s"/"%s"/g' %s`,
		moduleName, terraformDbVersion, actualDbVersion, fileName,
	)
	log.Printf("Sed command: %s", sedCmd)

	execCmd := exec.Command("/bin/sh", "-c", sedCmd)
	execCmd.Dir = tfDir
	if err := execCmd.Run(); err != nil {
		log.Printf("[ERROR] Failed to update hardcoded version with sed: %v", err)
		return false, err
	}
	return true, nil
}

func updateVersionVariable(actualDbVersion, terraformDbVersion, tfDir, varName string) (bool, string, error) {
	grepVarCmd := fmt.Sprintf("grep -l 'variable \"%s\"' *.tf", varName)
	cmd := exec.Command("/bin/sh", "-c", grepVarCmd)
	cmd.Dir = tfDir
	varOutput, err := cmd.Output()
	if err != nil {
		return false, "", fmt.Errorf("failed to find variable definition for %s: %w", varName, err)
	}
	varFile := strings.TrimSpace(string(varOutput))

	varSedCmd := fmt.Sprintf(
		"sed -i -E 's/default[[:space:]]*=[[:space:]]*\"%s\"/default = \"%s\"/' %s",
		terraformDbVersion, actualDbVersion, varFile)
	log.Printf("Sed command for variable update: %s", varSedCmd)

	cmd = exec.Command("/bin/sh", "-c", varSedCmd)
	cmd.Dir = tfDir
	if err := cmd.Run(); err != nil {
		log.Printf("[ERROR] Failed to update default value in variable file: %v", err)
		return false, "", fmt.Errorf("failed to update variable default: %w", err)
	}

	return true, varFile, nil
}
