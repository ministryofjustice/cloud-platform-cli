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
	// Find file(s) that contain the module declaration.
	grepCmd := exec.Command("/bin/sh", "-c", "grep -l 'module \""+moduleName+"\"' *")
	grepCmd.Dir = tfDir
	fileNameBytes, err := grepCmd.Output()
	if err != nil {
		return "", fmt.Errorf("git grep error: %v", err)
	}
	fileName := strings.TrimSpace(string(fileNameBytes))
	if fileName == "" {
		return "", fmt.Errorf("no file found for module %s", moduleName)
	}

	// Get a snippet of the file to locate the engine_version line.
	execLine := "grep -A 40 -n 'module \"" + moduleName + "\"' " + fileName
	lineCmd := exec.Command("/bin/sh", "-c", execLine)
	lineCmd.Dir = tfDir
	lineNumBytes, err := lineCmd.Output()
	if err != nil {
		return "", fmt.Errorf("error running grep for engine_version: %v", err)
	}
	lineNum := string(lineNumBytes)

	// Use a regex to find the line number that contains engine_version.
	dbEngine := regexp.MustCompile(`(\d+).*engine_version`)
	dbMatch := dbEngine.FindStringSubmatch(lineNum)
	if dbMatch == nil {
		return "", fmt.Errorf("no engine_version match found in %s", fileName)
	}

	// Prepare a sed command to update the version.
	splitTfDbVersion := strings.Split(terraformDbVersion, ".")
	escapedTfDbVersion := "\"" + splitTfDbVersion[0] + "\\." + splitTfDbVersion[1] + "\""
	sedCmd := exec.Command("/bin/sh", "-c", "sed -i'' -e '"+dbMatch[1]+"s/"+escapedTfDbVersion+"/\""+actualDbVersion+"\"/g' "+fileName)
	sedCmd.Dir = tfDir
	if err := sedCmd.Run(); err != nil {
		return "", fmt.Errorf("sed command failed: %v", err)
	}

	return fileName, nil
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
