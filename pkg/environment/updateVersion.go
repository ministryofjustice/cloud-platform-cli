package environment

import (
	"fmt"
	"log"
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

	var changedFiles []string

	hardcodeUpdated, err := updateVersionHardcode(moduleName, fileName, actualDbVersion, terraformDbVersion, tfDir, moduleBlock)
	if err == nil && hardcodeUpdated {
		changedFiles = append(changedFiles, fileName)
		return changedFiles, nil
	}

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
		return true, nil
	}
	return false, nil
}

func updateVersionVariable(fileName, moduleName, actualDbVersion, terraformDbVersion, tfDir string, moduleBlock string) (bool, string, error) {
	re := regexp.MustCompile(`(?i)db_engine_version\s*=\s*var\.([a-zA-Z0-9_\-]+)`)
	matches := re.FindStringSubmatch(moduleBlock)
	if len(matches) < 2 {
		log.Printf("No variable found for db_engine_version in module %q", moduleName)
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

	varSedCmd := fmt.Sprintf("sed -i -E 's/default[[:space:]]*=[[:space:]]*\"%s\"/default = \"%s\"/' %s",
		terraformDbVersion, actualDbVersion, varFile)
	updateVarCmd := exec.Command("/bin/sh", "-c", varSedCmd)
	updateVarCmd.Dir = tfDir
	if err := updateVarCmd.Run(); err != nil {
		return false, "", fmt.Errorf("failed to update variable default: %w", err)
	}

	return true, varFile, nil
}
