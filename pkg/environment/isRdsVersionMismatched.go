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

func IsRdsVersionMismatched(outputTerraform string) (*RdsVersionResults, error) {
	match, _ := regexp.MatchString(`(?i)Error: updating RDS .* InvalidParameterCombination:`, outputTerraform)
	if !match {
		return nil, errors.New("terraform is failing but it doesn't look like a rds version mismatch")
	}

	if !strings.Contains(outputTerraform, "Cannot upgrade") &&
		!strings.Contains(outputTerraform, "Cannot find upgrade path") {
		return nil, errors.New("terraform is failing but it doesn't look like a rds version mismatch")
	}

	versionRePrimary := regexp.MustCompile(`(?i)from ([^\s]+) (?:to|with requested version) ([^\s]+)`)
	versionReFallback := regexp.MustCompile(`(?i)Cannot find upgrade path from ([^\s,]+) to ([^\s,]+)[.,]?`)

	var versionMatches [][]string

	if strings.Contains(outputTerraform, "Cannot find upgrade path from") {
		versionMatches = versionReFallback.FindAllStringSubmatch(outputTerraform, -1)
		for i := range versionMatches {
			if len(versionMatches[i]) == 3 {
				versionMatches[i][1] = strings.TrimSuffix(versionMatches[i][1], ".")
				versionMatches[i][2] = strings.TrimSuffix(versionMatches[i][2], ".")
			}
		}
	} else {
		versionMatches = versionRePrimary.FindAllStringSubmatch(outputTerraform, -1)
	}

	moduleNameRe := regexp.MustCompile(`with module\.([^\s,]+?)\.aws_db_instance\.rds|aws_rds_cluster\.aurora,?`)
	moduleMatches := moduleNameRe.FindAllStringSubmatch(outputTerraform, -1)

	log.Printf("Raw versionMatches: %+v", versionMatches)
	log.Printf("Raw moduleMatches: %+v", moduleMatches)

	if len(versionMatches) == 0 {
		return nil, errors.New("terraform is failing but it doesn't look like a rds version mismatch")
	}

	sanitisedVersions := removeInputStr(versionMatches)
	sanitisedNames := removeInputStr(moduleMatches)

	log.Printf("sanitisedVersions: %+v", sanitisedVersions)
	log.Printf("sanitisedNames: %+v", sanitisedNames)

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
			log.Printf("Skipping invalid version pair (not length 2): %v", pair)
			return false
		}
		actual, desired := strings.Trim(pair[0], " ,."), strings.Trim(pair[1], " ,.")

		actInt, actErr := versionToInt(actual)
		desInt, desErr := versionToInt(desired)

		if actErr == nil && desErr == nil {
			if desInt >= actInt {
				log.Printf("Not a downgrade – desired (%s) >= actual (%s)", desired, actual)
				return false
			}
			log.Printf("Valid numeric downgrade: actual (%s) → desired (%s)", actual, desired)
			continue
		}

		actOracle, desOracle := extractOracleDate(actual), extractOracleDate(desired)
		if actOracle != 0 && desOracle != 0 {
			if desOracle >= actOracle {
				log.Printf("Not a downgrade (oracle) – desired (%s) >= actual (%s)", desired, actual)
				return false
			}
			log.Printf("Valid Oracle downgrade: actual (%s) → desired (%s)", actual, desired)
			continue
		}

		log.Printf("Failed to compare version pair – actual: %s, desired: %s", actual, desired)
		return false
	}
	return true
}

func extractOracleDate(version string) int {
	re := regexp.MustCompile(`ru-(\d{4})-(\d{2})\.rur-\d{4}-\d{2}\.r(\d)`) // Matches: ru-YYYY-MM.rur-YYYY-MM.rX
	match := re.FindStringSubmatch(version)
	if len(match) == 4 {
		dateStr := match[1] + match[2] + match[3] // e.g., "2025012"
		if val, err := strconv.Atoi(dateStr); err == nil {
			return val
		}
	}
	return 0
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
		ret = append(ret, inner[1:]...)
		outer = append(outer, ret)
	}
	return outer
}
