package environment

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func updateVersion(moduleName, actualDbVersion, terraformDbVersion, tfDir string) ([]string, error) {
	if _, err := os.Stat(tfDir); os.IsNotExist(err) {
		abs, _ := filepath.Abs(tfDir)
		return nil, fmt.Errorf("directory does not exist: %s (resolved to: %s)", tfDir, abs)
	}

	grepModule := exec.Command("/bin/sh", "-c", "grep -lr 'module \""+moduleName+"\"' *")
	grepModule.Dir = tfDir
	moduleFileBytes, err := grepModule.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to grep module reference: %w", err)
	}
	files := strings.Split(strings.TrimSpace(string(moduleFileBytes)), "\n")
	if len(files) == 0 {
		return nil, fmt.Errorf("no files found referencing module %s", moduleName)
	}
	fileName := files[0]
	changedFiles := []string{fileName}

	sedCmd := fmt.Sprintf("sed -i -e 's/db_engine_version *= *\"%s\"/db_engine_version = \"%s\"/g' %s", terraformDbVersion, actualDbVersion, fileName)
	hardcodedSed := exec.Command("/bin/sh", "-c", sedCmd)
	hardcodedSed.Dir = tfDir
	_ = hardcodedSed.Run()

	scanModuleBlock := exec.Command("/bin/sh", "-c", "grep -A 40 'module \""+moduleName+"\"' "+fileName)
	scanModuleBlock.Dir = tfDir
	scanOut, err := scanModuleBlock.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to scan module block: %w", err)
	}

	re := regexp.MustCompile(`(?i)db_engine_version *= *var\.([a-zA-Z0-9_\-]+)`)
	match := re.FindStringSubmatch(string(scanOut))
	if len(match) < 2 {
		return changedFiles, nil
	}
	varName := match[1]

	grepVar := exec.Command("/bin/sh", "-c", "grep -l 'variable \""+varName+"\"' *")
	grepVar.Dir = tfDir
	varFileBytes, err := grepVar.Output()
	if err != nil {
		return nil, fmt.Errorf("could not find variable definition for %s: %v", varName, err)
	}
	varFile := strings.TrimSpace(string(varFileBytes))

	varSedCmd := fmt.Sprintf("sed -i -e 's/default = \"%s\"/default = \"%s\"/' %s", terraformDbVersion, actualDbVersion, varFile)
	replaceDefault := exec.Command("/bin/sh", "-c", varSedCmd)
	replaceDefault.Dir = tfDir
	if err := replaceDefault.Run(); err != nil {
		return nil, fmt.Errorf("failed to update variable default: %v", err)
	}

	if !contains(changedFiles, varFile) {
		changedFiles = append(changedFiles, varFile)
	}

	return changedFiles, nil
}
