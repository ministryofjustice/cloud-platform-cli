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
	csvData := `foobar-ns-postgres-downgrade,Error: updating RDS DB Instance (cloud-platform-a): operation error RDS: ModifyDBInstance, api error InvalidParameterCombination: Cannot upgrade postgres from 14.13 to 14.7., with module.rds.aws_db_instance.rds,
foobar-ns-rds-non-downgrade,Error: updating RDS DB Instance (cloud-platform-b): operation error RDS: ModifyDBInstance, api error InvalidParameterCombination: Max storage size must be greater than storage size, with module.rds.aws_db_instance.rds,
foobar-ns-oracle-upgrade,Error: updating RDS DB Instance (cloud-platform-c): operation error RDS: ModifyDBInstance, api error InvalidParameterCombination: Cannot upgrade oracle-ee from 19.0.0.0.ru-2024-10.rur-2024-10.r1 to 19.0.0.0.ru-2025-01.rur-2025-01.r1, with module.oracle.aws_db_instance.rds,
foobar-ns-oracle-downgrade,Error: updating RDS DB Instance (cloud-platform-d): operation error RDS: ModifyDBInstance, api error InvalidParameterCombination: Cannot upgrade oracle-ee from 19.0.0.0.ru-2025-10.rur-2025-10.r1 to 19.0.0.0.ru-2024-01.rur-2024-01.r1, with module.oracle.aws_db_instance.rds,
foobar-ns-postgres-upgrade,Error: updating RDS DB Instance (cloud-platform-e): operation error RDS: ModifyDBInstance, api error InvalidParameterCombination: Cannot upgrade postgres from 14.7 to 14.13, with module.rds.aws_db_instance.rds,
foobar-ns-mariadb-downgrade,Error: updating RDS DB Instance (cloud-platform-f): operation error RDS: ModifyDBInstance, https response error StatusCode: 400, RequestID: xxx, api error InvalidParameterCombination: Cannot upgrade mariadb from 10.11.8 to 10.11.6, with module.rds.aws_db_instance.rds,
foobar-ns-mariadb-upgrade,Error: updating RDS DB Instance (cloud-platform-g): operation error RDS: ModifyDBInstance, https response error StatusCode: 400, RequestID: yyy, api error InvalidParameterCombination: Cannot upgrade mariadb from 10.11.8 to 10.11.9, with module.rds.aws_db_instance.rds,
foobar-ns-postrgres-same-ns,Error: updating RDS DB Instance (cloud-platform-h): operation error RDS: ModifyDBInstance, https response error StatusCode: 400, RequestID: zzz, api error InvalidParameterCombination: Cannot upgrade postgres from 16.6 to 16.4, with module.rds_2.aws_db_instance.rds,
foobar-ns-postrgres-same-ns,Error: updating RDS DB Instance (cloud-platform-i): operation error RDS: ModifyDBInstance, https response error StatusCode: 400, RequestID: aaa, api error InvalidParameterCombination: Cannot upgrade postgres from 16.6 to 16.4, with module.rds.aws_db_instance.rds,
foobar-ns-invalid-version-format,Error: updating RDS DB Instance (cloud-platform-j): api error InvalidParameterCombination: Cannot upgrade postgres from x.y to z.q, with module.rds.aws_db_instance.rds,
foobar-ns-missing-module-name,Error: updating RDS DB Instance (cloud-platform-k): api error InvalidParameterCombination: Cannot upgrade postgres from 13.6 to 12.4,
foobar-ns-oracle-same-ru-upgrade,Error: updating RDS DB Instance (cloud-platform-l): operation error RDS: ModifyDBInstance, api error InvalidParameterCombination: Cannot upgrade oracle-ee from 19.0.0.0.ru-2025-01.rur-2025-01.r1 to 19.0.0.0.ru-2025-01.rur-2025-01.r2, with module.oracle.aws_db_instance.rds,
foobar-ns-postgres-replica-downgrade,Error: updating RDS DB Instance (cloud-platform-m): operation error RDS: ModifyDBInstance, api error InvalidParameterCombination: Cannot upgrade postgres from 14.13 to 14.7., with module.rds_replica[0].aws_db_instance.rds,
`

	tmpFile := createTempCsv(t, csvData)
	defer os.Remove(tmpFile)

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

	tests := []struct {
		testDescription string
		ns              string
		wantErr         string
		wantSuccess     bool
	}{
		{
			testDescription: "GIVEN a downgrade mismatch for postgres THEN expect a VALID RESULT",
			ns:              "foobar-ns-postgres-downgrade",
			wantSuccess:     true,
		},
		{
			testDescription: "GIVEN a storage error THEN expect a INVALID RESULT due to non-mismatch RDS error",
			ns:              "foobar-ns-rds-non-downgrade",
			wantErr:         "terraform is failing but it doesn't look like a rds version mismatch",
		},
		{
			testDescription: "GIVEN an upgrade for oracle THEN expect a INVALID RESULT due to upgrade",
			ns:              "foobar-ns-oracle-upgrade",
			wantErr:         "terraform is failing, but it isn't trying to downgrade the RDS versions so it needs more investigation",
		},
		{
			testDescription: "GIVEN a downgrade mismatch for oracle THEN expect a VALID RESULT",
			ns:              "foobar-ns-oracle-downgrade",
			wantSuccess:     true,
		},
		{
			testDescription: "GIVEN an upgrade for postgres THEN expect a INVALID RESULT due to upgrade",
			ns:              "foobar-ns-postgres-upgrade",
			wantErr:         "terraform is failing, but it isn't trying to downgrade the RDS versions so it needs more investigation",
		},
		{
			testDescription: "GIVEN a downgrade mismatch for mariadb THEN expect a VALID RESULT",
			ns:              "foobar-ns-mariadb-downgrade",
			wantSuccess:     true,
		},
		{
			testDescription: "GIVEN an upgrade for mariadb THEN expect a INVALID RESULT due to upgrade",
			ns:              "foobar-ns-mariadb-upgrade",
			wantErr:         "terraform is failing, but it isn't trying to downgrade the RDS versions so it needs more investigation",
		},
		{
			testDescription: "GIVEN multiple downgrade errors in the same namespace THEN expect a VALID RESULT",
			ns:              "foobar-ns-postrgres-same-ns",
			wantSuccess:     true,
		},
		{
			testDescription: "GIVEN an invalid version format THEN expect a INVALID RESULT due to failed version parsing",
			ns:              "foobar-ns-invalid-version-format",
			wantErr:         "terraform is failing, but it isn't trying to downgrade the RDS versions so it needs more investigation",
		},
		{
			testDescription: "GIVEN a missing module name THEN expect a INVALID RESULT due to missing module and inconsistent version and module count",
			ns:              "foobar-ns-missing-module-name",
			wantErr:         "error: there is an inconsistent number of versions vs module names, there should be an even amount but we have 1 sets of versions and 0 module names",
		},
		{
			testDescription: "GIVEN an upgrade for oracle (same date but newer release version) THEN expect a INVALID RESULT due to upgrade",
			ns:              "foobar-ns-oracle-same-ru-upgrade",
			wantErr:         "terraform is failing, but it isn't trying to downgrade the RDS versions so it needs more investigation",
		},
		{
			testDescription: "GIVEN a downgrade mismatch for postgres replica THEN expect a VALID RESULT",
			ns:              "foobar-ns-postgres-replica-downgrade",
			wantSuccess:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testDescription, func(t *testing.T) {
			msgs := strings.Join(namespaceMap[tt.ns], "\n")
			_, err := environment.IsRdsVersionMismatched(msgs)

			if tt.wantSuccess {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			} else {
				if err == nil || err.Error() != tt.wantErr {
					t.Errorf("Expected error %q but got %v", tt.wantErr, err)
				}
			}
		})
	}
}
