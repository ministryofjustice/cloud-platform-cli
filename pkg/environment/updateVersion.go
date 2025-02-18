package environment

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

func updateVersion(moduleName, actualDbVersion, terraformDbVersion, tfDir string) (string, error) {
	grepCmd := exec.Command("/bin/sh", "-c", "grep -l 'module \""+moduleName+"\"' *")

	grepCmd.Dir = tfDir

	fileName, grepErr := grepCmd.Output()

	if grepErr != nil {
		return "", grepErr
	}

	execLine := "grep -A 40 -n 'module \"" + moduleName + "\"' " + string(fileName)
	lineCmd := exec.Command("/bin/sh", "-c", execLine)

	lineCmd.Dir = tfDir

	lineNum, lineErr := lineCmd.Output()

	dbEngine := regexp.MustCompile(`(\d+).*engine_version`)
	dbMatch := dbEngine.FindStringSubmatch(string(lineNum))

	if dbMatch == nil {
		return "", fmt.Errorf("no line match")
	}

	if lineErr != nil {
		return "", lineErr
	}

	splitTfDbVersion := strings.Split(terraformDbVersion, ".")

	escapedTfDbVersion := "\"" + splitTfDbVersion[0] + "\\." + splitTfDbVersion[1] + "\""

	sedCmd := exec.Command("/bin/sh", "-c", "sed -i'' -e '"+dbMatch[1]+"s/"+escapedTfDbVersion+"/\""+actualDbVersion+"\"/g' "+string(fileName))

	sedCmd.Dir = tfDir

	sedErr := sedCmd.Run()

	if sedErr != nil {
		return "", sedErr
	}

	return string(fileName), nil
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
