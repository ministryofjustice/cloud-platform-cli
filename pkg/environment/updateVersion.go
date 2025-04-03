package environment

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// updateVersion searches for the given module name in the target directory,
// and updates the Terraform version in the file from the reported (Terraform) version
// to the actual (desired) version.
func updateVersion(moduleName, actualDbVersion, terraformDbVersion, tfDir string) (string, error) {
	if _, err := os.Stat(tfDir); os.IsNotExist(err) {
		abs, _ := filepath.Abs(tfDir)
		return "", fmt.Errorf("directory does not exist: %s (resolved to: %s)", tfDir, abs)
	}
	grepModule := exec.Command("/bin/sh", "-c", "grep -lr 'module \""+moduleName+"\"' *")
	grepModule.Dir = tfDir
	moduleFileBytes, err := grepModule.Output()
	if err != nil {
		return "", fmt.Errorf("failed to grep module reference: %w", err)
	}
	fileName := strings.TrimSpace(string(moduleFileBytes))

	// Attempt to replace hardcoded db_engine_version
	sedCmd := fmt.Sprintf("sed -i'' -e 's/db_engine_version *= *\"%s\"/db_engine_version = \"%s\"/g' %s", terraformDbVersion, actualDbVersion, fileName)
	hardcodedSed := exec.Command("/bin/sh", "-c", sedCmd)
	hardcodedSed.Dir = tfDir
	_ = hardcodedSed.Run()

	// Check if a change occurred
	diffCheck := exec.Command("/bin/sh", "-c", "git diff --name-only "+fileName)
	diffCheck.Dir = tfDir
	diffOutput, _ := diffCheck.Output()
	if strings.TrimSpace(string(diffOutput)) != "" {
		return fileName, nil
	}

	// Fallback: handle variable case
	scanModuleBlock := exec.Command("/bin/sh", "-c", "grep -A 40 'module \""+moduleName+"\"' "+fileName)
	scanModuleBlock.Dir = tfDir
	scanOut, err := scanModuleBlock.Output()
	if err != nil {
		return "", fmt.Errorf("failed to scan module block: %w", err)
	}

	// Match variable reference (e.g., db_engine_version = var.some_version)
	re := regexp.MustCompile(`(?i)db_engine_version *= *var\.([a-zA-Z0-9_\-]+)`)
	match := re.FindStringSubmatch(string(scanOut))
	if match == nil || len(match) < 2 {
		return "", fmt.Errorf("could not find db_engine_version assignment in module block")
	}
	varName := match[1]

	// Find file that defines the variable
	grepVar := exec.Command("/bin/sh", "-c", "grep -l 'variable \""+varName+"\"' *")
	grepVar.Dir = tfDir
	varFileBytes, err := grepVar.Output()
	if err != nil {
		return "", fmt.Errorf("could not find variable definition for %s: %v", varName, err)
	}
	varFile := strings.TrimSpace(string(varFileBytes))

	// Replace default value in variable block
	varSedCmd := fmt.Sprintf("sed -i'' -e 's/default = \"%s\"/default = \"%s\"/' %s", terraformDbVersion, actualDbVersion, varFile)
	replaceDefault := exec.Command("/bin/sh", "-c", varSedCmd)
	replaceDefault.Dir = tfDir
	if err := replaceDefault.Run(); err != nil {
		return "", fmt.Errorf("failed to update variable default: %v", err)
	}

	return varFile, nil
}

// not need as used in apply pipeline
// func checkRdsAndUpdate(tfErr, tfDir string) (string, []string, error) {
// 	var filenames []string

// 	matches, rdsErr := IsRdsVersionMismatched(tfErr)
// 	versionDescription := "Fix Terraform RDS version drift. Here are the RDS version mismatches: \n\n"

// 	if rdsErr != nil {
// 		return "", nil, rdsErr
// 	}

// 	for i := 0; i < matches.TotalVersionMismatches; i++ {
// 		moduleName := matches.ModuleNames[i][0]
// 		actualDbVersion := matches.Versions[i][0]
// 		terraformDbVersion := matches.Versions[i][1]

// 		versionDescription += versionDescription + "downgrade from " + actualDbVersion + " to " + terraformDbVersion + "\n"

// 		file, updateErr := updateVersion(moduleName, actualDbVersion, terraformDbVersion, tfDir)
// 		filenames = append(filenames, file)
// 		if updateErr != nil {
// 			return "", nil, updateErr
// 		}
// 	}

// 	return versionDescription, filenames, nil
// }
