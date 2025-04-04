package environment_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/environment"
)

func createTempCsv(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.csv")

	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp CSV: %v", err)
	}
	return tmpFile
}

func TestRdsDriftChecker(t *testing.T) {
	csvData := `namespace,error_message
ns-downgrade,Error: updating RDS DB Instance (cloud-platform-x): operation error RDS: ModifyDBInstance, api error InvalidParameterCombination: Cannot upgrade postgres from 14.13 to 14.7.,   with module.rds.aws_db_instance.rds,
ns-storage,Error: updating RDS DB Instance (cloud-platform-y): operation error RDS: ModifyDBInstance, api error InvalidParameterCombination: Max storage size must be greater than storage size,   with module.rds.aws_db_instance.rds,
ns-oracle,Error: updating RDS DB Instance (cloud-platform-z): operation error RDS: ModifyDBInstance, api error InvalidParameterCombination: Cannot upgrade oracle-ee from 19.0.0.0.ru-2024-10.rur-2024-10.r1 to 19.0.0.0.ru-2025-01.rur-2025-01.r1,   with module.oracle.aws_db_instance.rds,
ns-oracle-2,Error: updating RDS DB Instance (cloud-platform-a): operation error RDS: ModifyDBInstance, api error InvalidParameterCombination: Cannot upgrade oracle-ee from 19.0.0.0.ru-2025-10.rur-2025-10.r1 to 19.0.0.0.ru-2024-01.rur-2024-01.r1,   with module.oracle.aws_db_instance.rds,
ns-upgrade,Error: updating RDS DB Instance (cloud-platform-w): operation error RDS: ModifyDBInstance, api error InvalidParameterCombination: Cannot upgrade postgres from 14.7 to 14.13,   with module.rds.aws_db_instance.rds,
ns-mariadb-1, Error: updating RDS DB Instance (cloud-platform-s): operation error RDS: ModifyDBInstance, https response error StatusCode: 400, RequestID: 9c326440-f357-4fc5-95b8-0fd5c33742c2, api error InvalidParameterCombination: Cannot upgrade mariadb from 10.11.8 to 10.11.6,   with module.rds.aws_db_instance.rds,
ns-mariadb-2, Error: updating RDS DB Instance (cloud-platform-s): operation error RDS: ModifyDBInstance, https response error StatusCode: 400, RequestID: 9c326440-f357-4fc5-95b8-0fd5c33742c2, api error InvalidParameterCombination: Cannot upgrade mariadb from 10.11.8 to 10.11.9,   with module.rds.aws_db_instance.rds,
`

	tmpFile := createTempCsv(t, csvData)

	fileBytes, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read test CSV: %v", err)
	}

	lines := strings.Split(string(fileBytes), "\n")[1:] // skip header

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, ",", 2)
		if len(parts) != 2 {
			t.Errorf("Invalid CSV line: %s", line)
			continue
		}
		ns := parts[0]
		msg := parts[1]

		t.Logf("[%s] ➤ Checking error: %s", ns, msg)
		res, err := environment.IsRdsVersionMismatched(msg)

		switch ns {
		case "ns-downgrade", "ns-oracle-2", "ns-mariadb-1":
			if err != nil || res == nil || res.TotalVersionMismatches == 0 {
				t.Errorf("[%s] ❌ Expected downgrade mismatch, got: %v / %v", ns, res, err)
			} else {
				t.Logf("[%s] ✅ Detected downgrade mismatch: %v", ns, res)
			}

		case "ns-storage":
			if err == nil || err.Error() != "terraform is failing but it doesn't look like a rds version mismatch" {
				t.Errorf("[%s] ❌ Expected non-mismatch error, got: %v", ns, err)
			} else {
				t.Logf("[%s] ✅ Correctly skipped non-mismatch error", ns)
			}

		case "ns-upgrade", "ns-oracle", "ns-mariadb-2":
			if err == nil || err.Error() != "terraform is failing, but it isn't trying to downgrade the RDS versions so it needs more investigation" {
				t.Errorf("[%s] ❌ Expected upgrade rejection, got: %v", ns, err)
			} else {
				t.Logf("[%s] ✅ Correctly identified upgrade instead of downgrade", ns)
			}

		default:
			t.Errorf("[%s] ⚠️ Unexpected namespace in test", ns)
		}
	}
}
