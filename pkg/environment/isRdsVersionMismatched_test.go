package environment_test

import (
	"encoding/csv"
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
	csvData := `foolbar-ns-postgres-downgrade,Error: updating RDS DB Instance (cloud-platform-x): operation error RDS: ModifyDBInstance, api error InvalidParameterCombination: Cannot upgrade postgres from 14.13 to 14.7., with module.rds.aws_db_instance.rds,
foolbar-ns-rds-non-downgrade,Error: updating RDS DB Instance (cloud-platform-y): operation error RDS: ModifyDBInstance, api error InvalidParameterCombination: Max storage size must be greater than storage size, with module.rds.aws_db_instance.rds,
foolbar-ns-oracle-upgrade,Error: updating RDS DB Instance (cloud-platform-z): operation error RDS: ModifyDBInstance, api error InvalidParameterCombination: Cannot upgrade oracle-ee from 19.0.0.0.ru-2024-10.rur-2024-10.r1 to 19.0.0.0.ru-2025-01.rur-2025-01.r1, with module.oracle.aws_db_instance.rds,
foolbar-ns-oracle-downgrade,Error: updating RDS DB Instance (cloud-platform-a): operation error RDS: ModifyDBInstance, api error InvalidParameterCombination: Cannot upgrade oracle-ee from 19.0.0.0.ru-2025-10.rur-2025-10.r1 to 19.0.0.0.ru-2024-01.rur-2024-01.r1, with module.oracle.aws_db_instance.rds,
foolbar-ns-postgres-upgrade,Error: updating RDS DB Instance (cloud-platform-w): operation error RDS: ModifyDBInstance, api error InvalidParameterCombination: Cannot upgrade postgres from 14.7 to 14.13, with module.rds.aws_db_instance.rds,
foolbar-ns-mariadb-downgrade,Error: updating RDS DB Instance (cloud-platform-s): operation error RDS: ModifyDBInstance, https response error StatusCode: 400, RequestID: xxx, api error InvalidParameterCombination: Cannot upgrade mariadb from 10.11.8 to 10.11.6, with module.rds.aws_db_instance.rds,
foolbar-ns-mariadb-upgrade,Error: updating RDS DB Instance (cloud-platform-s): operation error RDS: ModifyDBInstance, https response error StatusCode: 400, RequestID: yyy, api error InvalidParameterCombination: Cannot upgrade mariadb from 10.11.8 to 10.11.9, with module.rds.aws_db_instance.rds,
foolbar-ns-postrgres-same-ns,Error: updating RDS DB Instance (cloud-platform-01152bdef9e4faaf): operation error RDS: ModifyDBInstance, https response error StatusCode: 400, RequestID: zzz, api error InvalidParameterCombination: Cannot upgrade postgres from 16.6 to 16.4, with module.rds_2.aws_db_instance.rds,
foolbar-ns-postrgres-same-ns,Error: updating RDS DB Instance (cloud-platform-35c7710fbe6aa229): operation error RDS: ModifyDBInstance, https response error StatusCode: 400, RequestID: aaa, api error InvalidParameterCombination: Cannot upgrade postgres from 16.6 to 16.4, with module.rds.aws_db_instance.rds,
foolbar-ns-invalid-version-format,Error: updating RDS DB Instance: api error InvalidParameterCombination: Cannot upgrade postgres from x.y to z.q, with module.rds.aws_db_instance.rds,
foolbar-ns-missing-module-name,Error: updating RDS DB Instance: api error InvalidParameterCombination: Cannot upgrade postgres from 13.6 to 12.4,
foolbar-ns-oracle-same-ru-upgrade,Error: updating RDS DB Instance (cloud-platform-test): operation error RDS: ModifyDBInstance, api error InvalidParameterCombination: Cannot upgrade oracle-ee from 19.0.0.0.ru-2025-01.rur-2025-01.r1 to 19.0.0.0.ru-2025-01.rur-2025-01.r2, with module.oracle.aws_db_instance.rds,
`

	tmpFile := createTempCsv(t, csvData)

	file, err := os.Open(tmpFile)
	if err != nil {
		t.Fatalf("Failed to open temp CSV: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to parse CSV: %v", err)
	}

	namespaceMap := make(map[string][]string)
	for _, record := range records {
		if len(record) < 2 {
			t.Logf("Skipping incomplete record: %v", record)
			continue
		}
		namespace := strings.TrimSpace(record[0])
		errorMsg := strings.TrimSpace(strings.Join(record[1:], " "))
		namespaceMap[namespace] = append(namespaceMap[namespace], errorMsg)
	}

	for ns, msgs := range namespaceMap {
		combined := strings.Join(msgs, "\n")
		t.Logf("[%s] ➤ Checking combined error:\n%s", ns, combined)

		res, err := environment.IsRdsVersionMismatched(combined)

		switch ns {
		case "foolbar-ns-postgres-downgrade", "foolbar-ns-oracle-downgrade", "foolbar-ns-mariadb-downgrade", "foolbar-ns-postrgres-same-ns":
			if err != nil || res == nil || res.TotalVersionMismatches == 0 {
				t.Errorf("[%s] Expected downgrade mismatch, got: %v / %v", ns, res, err)
			}

		case "foolbar-ns-rds-non-downgrade":
			want := "terraform is failing but it doesn't look like a rds version mismatch"
			if err == nil || err.Error() != want {
				t.Errorf("[%s] Expected non-mismatch error, got: %v", ns, err)
			}

		case "foolbar-ns-postgres-upgrade", "foolbar-ns-oracle-upgrade", "foolbar-ns-mariadb-upgrade", "foolbar-ns-oracle-same-ru-upgrade":
			want := "terraform is failing, but it isn't trying to downgrade the RDS versions so it needs more investigation"
			if err == nil || err.Error() != want {
				t.Errorf("[%s] Expected upgrade rejection, got: %v", ns, err)
			}

		case "foolbar-ns-invalid-version-format":
			want := "terraform is failing, but it isn't trying to downgrade the RDS versions so it needs more investigation"
			if err == nil || err.Error() != want {
				t.Errorf("[%s] Expected failure on invalid version comparison, got: %v", ns, err)
			}

		case "foolbar-ns-missing-module-name":
			if err == nil {
				t.Errorf("[%s] Expected error due to missing module name, but got: %v", ns, res)
			}

		default:
			t.Errorf("[%s] Unexpected namespace in test", ns)
		}
	}
}
