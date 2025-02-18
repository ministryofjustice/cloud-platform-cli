package environment

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type RdsVersionResults struct {
	Versions               [][]string
	ModuleNames            [][]string
	TotalVersionMismatches int
}

func IsRdsVersionMismatched(outputTerraform string) (*RdsVersionResults, error) {
	match, _ := regexp.MatchString("Error: updating RDS .* api error InvalidParameterCombination:.* from .* (?:to|with requested version) .*", outputTerraform)

	if !match {
		return nil, errors.New("terraform is failing but it doesn't look like a rds version mismatch")
	}

	versionRe := regexp.MustCompile(`from (?P<actual_db_version>\d+\.\d+) (?:to|with requested version) (?P<terraform_db_version>\d+\.\d+)`)

	moduleNameRe := regexp.MustCompile(`with module\.(.+)\.(?:aws_db_instance\.rds|aws_rds_cluster\.aurora),`)

	moduleMatches := moduleNameRe.FindAllStringSubmatch(outputTerraform, -1)
	versionMatches := versionRe.FindAllStringSubmatch(outputTerraform, -1)

	sanitisedVersions := removeInputStr(versionMatches)
	sanitisedNames := removeInputStr(moduleMatches)

	if !checkVersionDowngrade(sanitisedVersions) {
		return nil, errors.New("terraform is failing, but it isn't trying to downgrade the RDS versions so it needs more investigation")
	}

	if len(sanitisedVersions) != len(sanitisedNames) {
		return nil, fmt.Errorf("error: there is an inconistent number of versions vs module names, there should be an even amount but we have %d sets of versions and %d module names", len(sanitisedVersions), len(sanitisedNames))
	}

	return &RdsVersionResults{
		sanitisedVersions,
		sanitisedNames,
		len(sanitisedVersions),
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
		ret = append(ret, inner[1:]...)
		outer = append(outer, ret)
	}
	return outer
}
