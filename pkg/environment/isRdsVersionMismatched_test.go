package environment_test

import (
	"os"
	"strings"
	"testing"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/environment"
)

func createTempCsv(t *testing.T, content string) string {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "rds-drift-test-*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}
	return tmpFile.Name()
}

func TestIsRdsVersionMismatched(t *testing.T) {
	csvData := `namespace,error_message
foolbar-ns-postgres-downgrade,Error: updating RDS DB Instance (cloud-platform-x): operation error RDS: ModifyDBInstance, api error InvalidParameterCombination: Cannot upgrade postgres from 14.13 to 14.7.,   with module.rds.aws_db_instance.rds,
foolbar-ns-rds-non-downgrade,Error: updating RDS DB Instance (cloud-platform-y): operation error RDS: ModifyDBInstance, api error InvalidParameterCombination: Max storage size must be greater than storage size,   with module.rds.aws_db_instance.rds,
foolbar-ns-oracle-upgrade,Error: updating RDS DB Instance (cloud-platform-z): operation error RDS: ModifyDBInstance, api error InvalidParameterCombination: Cannot upgrade oracle-ee from 19.0.0.0.ru-2024-10.rur-2024-10.r1 to 19.0.0.0.ru-2025-01.rur-2025-01.r1,   with module.oracle.aws_db_instance.rds,
foolbar-ns-oracle-downgrade,Error: updating RDS DB Instance (cloud-platform-a): operation error RDS: ModifyDBInstance, api error InvalidParameterCombination: Cannot upgrade oracle-ee from 19.0.0.0.ru-2025-10.rur-2025-10.r1 to 19.0.0.0.ru-2024-01.rur-2024-01.r1,   with module.oracle.aws_db_instance.rds,
foolbar-ns-postgres-upgrade,Error: updating RDS DB Instance (cloud-platform-w): operation error RDS: ModifyDBInstance, api error InvalidParameterCombination: Cannot upgrade postgres from 14.7 to 14.13,   with module.rds.aws_db_instance.rds,
foolbar-ns-mariadb-downgrade,Error: updating RDS DB Instance (cloud-platform-s): operation error RDS: ModifyDBInstance, https response error StatusCode: 400, RequestID: 9c326440-f357-4fc5-95b8-0fd5c33742c2, api error InvalidParameterCombination: Cannot upgrade mariadb from 10.11.8 to 10.11.6,   with module.rds.aws_db_instance.rds,
foolbar-ns-mariadb-upgrade,Error: updating RDS DB Instance (cloud-platform-s): operation error RDS: ModifyDBInstance, https response error StatusCode: 400, RequestID: 9c326440-f357-4fc5-95b8-0fd5c33742c2, api error InvalidParameterCombination: Cannot upgrade mariadb from 10.11.8 to 10.11.9,   with module.rds.aws_db_instance.rds,
foolbar-ns-postrgres-same-ns,Error: updating RDS DB Instance (cloud-platform-01152bdef9e4faaf): operation error RDS: ModifyDBInstance, https response error StatusCode: 400, RequestID: 3693bb6d-65f7-4024-a068-fc1af28e1900, api error InvalidParameterCombination: Cannot upgrade postgres from 16.6 to 16.4,   with module.rds_2.aws_db_instance.rds,
foolbar-ns-postrgres-same-ns,Error: updating RDS DB Instance (cloud-platform-35c7710fbe6aa229): operation error RDS: ModifyDBInstance, https response error StatusCode: 400, RequestID: 093a8d4b-d60c-4e6c-94ce-a3121c6e8e53, api error InvalidParameterCombination: Cannot upgrade postgres from 16.6 to 16.4,   with module.rds.aws_db_instance.rds,
`

	tmpFile := createTempCsv(t, csvData)

	fileBytes, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read test CSV: %v", err)
	}

	lines := strings.Split(string(fileBytes), "\n")[1:] // skip header
	namespaceCombined := make(map[string]string)

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, ",", 2)
		if len(parts) != 2 {
			t.Errorf("Invalid CSV line: %s", line)
			continue
		}
		ns := strings.TrimSpace(parts[0])
		msg := strings.TrimSpace(parts[1])
		namespaceCombined[ns] += msg + "\n"
	}

	for ns, msg := range namespaceCombined {
		t.Logf("[%s] ➤ Checking error: %s", ns, msg)
		res, err := environment.IsRdsVersionMismatched(msg)

		switch ns {
		case "foolbar-ns-postgres-downgrade", "foolbar-ns-oracle-downgrade", "foolbar-ns-mariadb-downgrade", "foolbar-ns-postrgres-same-ns":
			if err != nil || res == nil || res.TotalVersionMismatches == 0 {
				t.Errorf("[%s] ❌ Expected downgrade mismatch, got: %v / %v", ns, res, err)
			} else {
				t.Logf("[%s] ✅ Detected downgrade mismatch: %v", ns, res)
			}

		case "foolbar-ns-rds-non-downgrade":
			if err == nil || err.Error() != "terraform is failing but it doesn't look like a rds version mismatch" {
				t.Errorf("[%s] ❌ Expected non-mismatch error, got: %v", ns, err)
			} else {
				t.Logf("[%s] ✅ Correctly skipped non-mismatch error", ns)
			}

		case "foolbar-ns-postgres-upgrade", "foolbar-ns-oracle-upgrade", "foolbar-ns-mariadb-upgrade":
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
