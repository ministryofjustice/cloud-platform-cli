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
	// First check it's an RDS InvalidParameterCombination error
	match, _ := regexp.MatchString(`(?i)Error: updating RDS .* InvalidParameterCombination:`, outputTerraform)
	if !match {
		return nil, errors.New("terraform is failing but it doesn't look like a rds version mismatch")
	}

	// Now make sure it's specifically about version mismatch (not storage, etc)
	if !strings.Contains(outputTerraform, "Cannot upgrade") &&
		!strings.Contains(outputTerraform, "Cannot find upgrade path") {
		return nil, errors.New("terraform is failing but it doesn't look like a rds version mismatch")
	}

	// Define regex patterns
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

	// Module name regex
	moduleNameRe := regexp.MustCompile(`with module\.([^\s,]+?)\.aws_db_instance\.rds,?`)
	moduleMatches := moduleNameRe.FindAllStringSubmatch(outputTerraform, -1)

	log.Printf("DEBUG: Raw versionMatches: %+v", versionMatches)
	log.Printf("DEBUG: Raw moduleMatches: %+v", moduleMatches)

	// If we found no version matches, it's not a version mismatch
	if len(versionMatches) == 0 {
		return nil, errors.New("terraform is failing but it doesn't look like a rds version mismatch")
	}

	sanitisedVersions := removeInputStr(versionMatches)
	sanitisedNames := removeInputStr(moduleMatches)

	log.Printf("DEBUG: sanitisedVersions: %+v", sanitisedVersions)
	log.Printf("DEBUG: sanitisedNames: %+v", sanitisedNames)

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
			log.Printf("DEBUG: Skipping invalid version pair (not length 2): %v", pair)
			return false
		}
		actual, desired := strings.Trim(pair[0], " ,."), strings.Trim(pair[1], " ,.")

		// Try numeric version parsing (works for postgres, mariadb, mysql, etc.)
		actInt, actErr := versionToInt(actual)
		desInt, desErr := versionToInt(desired)

		if actErr == nil && desErr == nil {
			if desInt >= actInt {
				log.Printf("DEBUG: Not a downgrade – desired (%s) >= actual (%s)", desired, actual)
				return false
			}
			log.Printf("DEBUG: Valid numeric downgrade (postgres/mariadb): actual (%s) → desired (%s)", actual, desired)
			continue
		}

		// Oracle fallback
		actOracle, desOracle := extractOracleDate(actual), extractOracleDate(desired)
		if actOracle != 0 && desOracle != 0 {
			if desOracle >= actOracle {
				log.Printf("DEBUG: Not a downgrade (oracle) – desired (%s) >= actual (%s)", desired, actual)
				return false
			}
			log.Printf("DEBUG: Valid Oracle downgrade: actual (%s) → desired (%s)", actual, desired)
			continue
		}

		log.Printf("DEBUG: Failed to compare version pair – actual: %s, desired: %s", actual, desired)
		return false
	}
	return true
}

func extractOracleDate(version string) int {
	re := regexp.MustCompile(`ru-(\d{4})-(\d{2})`)
	match := re.FindStringSubmatch(version)
	if len(match) == 3 {
		dateStr := match[1] + match[2] // "202501"
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
