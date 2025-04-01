package environment

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

// RdsVersionResults holds the extracted version and module name information.
type RdsVersionResults struct {
	Versions               [][]string
	ModuleNames            [][]string
	TotalVersionMismatches int
}

func IsRdsVersionMismatched(outputTerraform string) (*RdsVersionResults, error) {
	// Check for a version-downgrade error message.
	match, _ := regexp.MatchString("Cannot upgrade postgres from \\d+\\.\\d+ to \\d+\\.\\d+", outputTerraform)
	if !match {
		return nil, errors.New("terraform is failing but it doesn't look like a rds version mismatch")
	}

	// Regular expression to capture version numbers.
	versionRe := regexp.MustCompile(`from (\d+\.\d+) to (\d+\.\d+)`)
	// Modified regex to capture module name.
	// This will match "with module." then capture any characters until the first period.
	moduleNameRe := regexp.MustCompile(`with module\.([^\s,]+?)\.aws_db_instance\.rds,?`)

	// Find all matches.
	moduleMatches := moduleNameRe.FindAllStringSubmatch(outputTerraform, -1)
	versionMatches := versionRe.FindAllStringSubmatch(outputTerraform, -1)

	// Debug: log raw matches.
	log.Printf("DEBUG: versionMatches: %+v", versionMatches)
	log.Printf("DEBUG: moduleMatches: %+v", moduleMatches)

	sanitisedVersions := removeInputStr(versionMatches)
	sanitisedNames := removeInputStr(moduleMatches)

	// Debug: log sanitised results.
	log.Printf("DEBUG: sanitisedVersions: %+v", sanitisedVersions)
	log.Printf("DEBUG: sanitisedNames: %+v", sanitisedNames)

	// Validate that the versions indicate a downgrade.
	if !checkVersionDowngrade(sanitisedVersions) {
		return nil, errors.New("terraform is failing, but it isn't trying to downgrade the RDS versions so it needs more investigation")
	}

	// Ensure we have the same count for versions and module names.
	if len(sanitisedVersions) != len(sanitisedNames) {
		return nil, fmt.Errorf("error: there is an inconistent number of versions vs module names, there should be an even amount but we have %d sets of versions and %d module names", len(sanitisedVersions), len(sanitisedNames))
	}

	return &RdsVersionResults{
		Versions:               sanitisedVersions,
		ModuleNames:            sanitisedNames,
		TotalVersionMismatches: len(sanitisedVersions),
	}, nil
}

func checkVersionDowngrade(versions [][]string) bool {
	isValid := true
	for _, inner := range versions {
		if len(inner) == 2 {
			splitAcc := strings.Split(inner[0], ".")
			splitTf := strings.Split(inner[1], ".")
			adjustedAcc := strings.Join(splitAcc, "")
			adjustedTf := strings.Join(splitTf, "")
			acc, accErr := strconv.ParseInt(adjustedAcc, 0, 64)
			tf, tfErr := strconv.ParseInt(adjustedTf, 0, 64)
			isUpgrade := tf > acc
			if accErr != nil || tfErr != nil || isUpgrade {
				isValid = false
				break
			}
		} else {
			isValid = false
			break
		}
	}
	return isValid
}

func removeInputStr(res [][]string) [][]string {
	outer := make([][]string, 0)
	for _, inner := range res {
		ret := make([]string, 0)
		// Remove the full match (first element) and keep the captured groups.
		ret = append(ret, inner[1:]...)
		outer = append(outer, ret)
	}
	return outer
}
