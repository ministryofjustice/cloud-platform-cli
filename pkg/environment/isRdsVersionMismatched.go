package environment

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

type RdsVersionResults struct {
	Versions               [][]string
	ModuleNames            [][]string
	TotalVersionMismatches int
}

func IsRdsVersionMismatched(csvErr string) (*RdsVersionResults, error) {
	match, _ := regexp.MatchString("Error: updating RDS .* api error InvalidParameterCombination:.* from .* (?:to|with requested version) .*", csvErr)
	if !match {
		return nil, errors.New("terraform is failing but it doesn't look like a rds version mismatch")
	}

	versionRegex := regexp.MustCompile(`(?i)from ([^\s]+) (?:to|with requested version) ([^\s]+)`)
	versionMatches := versionRegex.FindAllStringSubmatch(csvErr, -1)
	log.Printf("Raw versionMatches: %+v", versionMatches)

	moduleNameRe := regexp.MustCompile(`with module\.([^\s,]+?)\.aws_db_instance\.rds|aws_rds_cluster\.aurora,?`)
	moduleMatches := moduleNameRe.FindAllStringSubmatch(csvErr, -1)

	sanitisedVersions := removeInputStr(versionMatches)
	sanitisedNames := removeInputStr(moduleMatches)

	log.Printf("Sanitised Versions: %+v", sanitisedVersions)
	log.Printf("Sanitised Module Names: %+v", sanitisedNames)

	if !checkVersionDowngrade(sanitisedVersions) {
		return nil, errors.New("terraform is failing, but it isn't trying to downgrade the RDS versions so it needs more investigation")
	}

	if len(sanitisedVersions) != len(sanitisedNames) {
		return nil, fmt.Errorf("error: there is an inconsistent number of versions vs module names, there should be an even amount but we have %d sets of versions and %d module names", len(sanitisedVersions), len(sanitisedNames))
	}

	return &RdsVersionResults{
		Versions:               sanitisedVersions,
		ModuleNames:            sanitisedNames,
		TotalVersionMismatches: len(sanitisedVersions),
	}, nil
}

func checkVersionDowngrade(versions [][]string) bool {
	for _, pair := range versions {
		if len(pair) != 2 {
			log.Printf("Skipping invalid version pair: %v", pair)
			return false
		}
		actual, desired := strings.Trim(pair[0], " ,."), strings.Trim(pair[1], " ,.")

		actInt, actErr := versionToInt(actual)
		desInt, desErr := versionToInt(desired)

		if actErr == nil && desErr == nil {
			if desInt >= actInt {
				return false
			}
			continue
		}

		return false
	}
	return true
}

func versionToInt(version string) (int64, error) {
	parts := strings.FieldsFunc(version, func(r rune) bool {
		return !(r >= '0' && r <= '9')
	})
	return strconv.ParseInt(strings.Join(parts, ""), 10, 64)
}

func removeInputStr(res [][]string) [][]string {
	outer := make([][]string, 0)
	for _, inner := range res {
		ret := make([]string, 0)
		for _, val := range inner[1:] {
			clean := strings.Trim(val, " .,")
			ret = append(ret, clean)
		}
		outer = append(outer, ret)
	}
	return outer
}
