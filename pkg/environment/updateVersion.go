package environment

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

func UpdateVersion(moduleName, actualDbVersion, terraformDbVersion, tfDir string) (string, error) {
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

	dbEngine := regexp.MustCompile(`(\d+).*db_engine_version`)
	dbMatch := dbEngine.FindStringSubmatch(string(lineNum))

	if dbMatch == nil {
		return "", fmt.Errorf("no line match")
	}

	if lineErr != nil {
		return "", lineErr
	}

	splitTfDbVersion := strings.Split(terraformDbVersion, ".")

	escapedTfDbVersion := "\"" + splitTfDbVersion[0] + "\\." + splitTfDbVersion[1] + "\""

	sedCmd := exec.Command("/bin/sh", "-c", "sed -i '' -e '"+dbMatch[1]+"s/"+escapedTfDbVersion+"/\""+actualDbVersion+"\"/g' "+string(fileName))

	sedCmd.Dir = tfDir

	sedErr := sedCmd.Run()

	if sedErr != nil {
		return "", sedErr
	}

	return string(fileName), nil
}
