package environment

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

func updateVersion(moduleName, actualDbVersion, terraformDbVersion, tfDir string) ([]string, error) {
	grepCmd := fmt.Sprintf("grep -l 'module \"%s\"' *.tf", moduleName)
	cmd := exec.Command("/bin/sh", "-c", grepCmd)
	cmd.Dir = tfDir
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to find module file: %w", err)
	}
	fileName := strings.TrimSpace(string(output))

	scanCmd := fmt.Sprintf("grep -A 40 'module \"%s\"' %s", moduleName, fileName)
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

	var changedFiles []string

	if strings.Contains(dbValue, "var.") {
		varName := strings.TrimSpace(strings.TrimPrefix(dbValue, "var."))
		updated, varFile, err := updateVersionVariable(actualDbVersion, terraformDbVersion, tfDir, varName)
		if err != nil {
			return nil, err
		}
		if updated && varFile != "" {
			changedFiles = append(changedFiles, varFile)
		}
	} else {
		updated, err := updateVersionHardcode(moduleName, fileName, actualDbVersion, terraformDbVersion, tfDir)
		if err != nil {
			return nil, err
		}
		if updated {
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
	execCmd := exec.Command("/bin/sh", "-c", sedCmd)
	execCmd.Dir = tfDir
	if err := execCmd.Run(); err != nil {
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
	cmd = exec.Command("/bin/sh", "-c", varSedCmd)
	cmd.Dir = tfDir
	if err := cmd.Run(); err != nil {
		return false, "", fmt.Errorf("failed to update variable default: %w", err)
	}

	return true, varFile, nil
}
