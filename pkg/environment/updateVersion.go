package environment

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// updateVersion searches for the given module name in the target directory,
// and updates the Terraform version in the file from the reported (Terraform) version
// to the actual (desired) version.
func updateVersion(moduleName, actualDbVersion, terraformDbVersion, tfDir string) (string, error) {
	// Try find a hardcoded db_engine_version assignment
	grepHardcoded := exec.Command("/bin/sh", "-c", "grep -lr 'module \""+moduleName+"\"' *")
	grepHardcoded.Dir = tfDir
	fileBytes, err := grepHardcoded.Output()
	if err != nil {
		return "", fmt.Errorf("grep module error: %v", err)
	}
	fileName := strings.TrimSpace(string(fileBytes))

	// Attempt to directly replace hardcoded engine version first
	hardcodeSed := exec.Command("/bin/sh", "-c", "sed -i'' -e 's/db_engine_version *= *\""+terraformDbVersion+"\"/db_engine_version = \""+actualDbVersion+"\"/g' "+fileName)
	hardcodeSed.Dir = tfDir
	hardcodeSed.Run()

	// Check if file actually changed (simple diff check)
	diffCmd := exec.Command("/bin/sh", "-c", "git diff --name-only "+fileName)
	diffCmd.Dir = tfDir
	diffOut, _ := diffCmd.Output()
	if strings.TrimSpace(string(diffOut)) != "" {
		return fileName, nil
	}

	// Otherwise, it's probably using a variable
	scanCmd := exec.Command("/bin/sh", "-c", "grep -A 40 'module \""+moduleName+"\"' "+fileName)
	scanCmd.Dir = tfDir
	scanOut, _ := scanCmd.Output()
	re := regexp.MustCompile(`db_engine_version *= *var\.([a-zA-Z0-9_\-]+)`)
	match := re.FindStringSubmatch(string(scanOut))
	if match == nil {
		return "", fmt.Errorf("could not find db_engine_version assignment in module block")
	}
	varName := match[1]

	// Look for the variable definition in the directory
	grepVar := exec.Command("/bin/sh", "-c", "grep -l 'variable \""+varName+"\"' *")
	grepVar.Dir = tfDir
	varFileBytes, err := grepVar.Output()
	if err != nil {
		return "", fmt.Errorf("could not find variable definition for %s: %v", varName, err)
	}
	varFile := strings.TrimSpace(string(varFileBytes))

	// Update the default value
	sedVar := exec.Command("/bin/sh", "-c", "sed -i'' -e 's/default = \""+terraformDbVersion+"\"/default = \""+actualDbVersion+"\"/' "+varFile)
	sedVar.Dir = tfDir
	if err := sedVar.Run(); err != nil {
		return "", fmt.Errorf("failed to update variable default: %v", err)
	}

	return varFile, nil
}

func checkRdsAndUpdate(tfErr, tfDir string) (string, []string, error) {
	var filenames []string

	matches, rdsErr := IsRdsVersionMismatched(tfErr)
	versionDescription := "Fix Terraform RDS version drift. Here are the RDS version mismatches: \n\n"

	if rdsErr != nil {
		return "", nil, rdsErr
	}

	for i := 0; i < matches.TotalVersionMismatches; i++ {
		moduleName := matches.ModuleNames[i][0]
		actualDbVersion := matches.Versions[i][0]
		terraformDbVersion := matches.Versions[i][1]

		versionDescription += versionDescription + "downgrade from " + actualDbVersion + " to " + terraformDbVersion + "\n"

		file, updateErr := updateVersion(moduleName, actualDbVersion, terraformDbVersion, tfDir)

		filenames = append(filenames, file)

		if updateErr != nil {
			return "", nil, updateErr
		}

	}

	return versionDescription, filenames, nil
}
