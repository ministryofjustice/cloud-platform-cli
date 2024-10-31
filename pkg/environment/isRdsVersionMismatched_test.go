package environment_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/environment"
)

var mismatchOutput = `2024/09/13 07:08:05 Running Terraform Apply for namespace: foobar

FATA[2623] error running terraform on namespace foobar: unable to apply Terraform: exit status 1


Error: updating RDS DB Instance (cloud-platform-123456789): operation error RDS: ModifyDBInstance, https response error StatusCode: 400, RequestID: xxxxxxxx-yyyyyy-zzzzzzzzz-, api error InvalidParameterCombination: Cannot upgrade postgres from 14.12 to 14.10


  with module.rds_example_module_name.aws_db_instance.rds,

  on .terraform/modules/foobar/main.tf line 166, in resource "aws_db_instance" "rds":

 166: resource "aws_db_instance" "rds" {`

var multipleMismatchOutput = `2024/09/13 07:08:05 Running Terraform Apply for namespace: foobar

FATA[2623] error running terraform on namespace foobar: unable to apply Terraform: exit status 1


Error: updating RDS DB Instance (cloud-platform-123456789): operation error RDS: ModifyDBInstance, https response error StatusCode: 400, RequestID: xxxxxxxx-yyyyyy-zzzzzzzzz-, api error InvalidParameterCombination: Cannot upgrade postgres from 14.12 to 14.10


  with module.rds_example_module_name.aws_db_instance.rds,

  on .terraform/modules/foobar/main.tf line 166, in resource "aws_db_instance" "rds":

 166: resource "aws_db_instance" "rds" {


Error: updating RDS DB Instance (cloud-platform-xxx): operation error RDS: ModifyDBInstance, https response error StatusCode: 400, RequestID: xxxxxxxx-yyyyyy-zzzzzzzzz-, api error InvalidParameterCombination: Cannot upgrade postgres from 16.18 to 16.16


  with module.rds_example_module_name_2.aws_db_instance.rds,

  on .terraform/modules/foobar/main.tf line 166, in resource "aws_db_instance" "rds":

 166: resource "aws_db_instance" "rds" {`

var nomatchOutput = "some random string not containing the right error"

var isVersionUpgradeMismatch = `2024/09/13 07:08:05 Running Terraform Apply for namespace: foobar

FATA[2623] error running terraform on namespace foobar: unable to apply Terraform: exit status 1


Error: updating RDS DB Instance (cloud-platform-123456789): operation error RDS: ModifyDBInstance, https response error StatusCode: 400, RequestID: xxxxxxxx-yyyyyy-zzzzzzzzz-, api error InvalidParameterCombination: Cannot upgrade postgres from 14.12 to 14.19


  with module.rds_example_module_name.aws_db_instance.rds,

  on .terraform/modules/foobar/main.tf line 166, in resource "aws_db_instance" "rds":

 166: resource "aws_db_instance" "rds" {`

var inconsistentOutput = `2024/09/13 07:08:05 Running Terraform Apply for namespace: foobar

FATA[2623] error running terraform on namespace foobar: unable to apply Terraform: exit status 1


Error: updating RDS DB Instance (cloud-platform-123456789): operation error RDS: ModifyDBInstance, https response error StatusCode: 400, RequestID: xxxxxxxx-yyyyyy-zzzzzzzzz-, api error InvalidParameterCombination: Cannot upgrade postgres from 14.12 to 14.10


  with module.rds_example_module_name.aws_db_instance.rds,
  with module.rds_example_module_name_2.aws_db_instance.rds,
  with module.rds_example_module_name_3.aws_db_instance.rds,

  on .terraform/modules/foobar/main.tf line 166, in resource "aws_db_instance" "rds":

 166: resource "aws_db_instance" "rds" {`

func TestIsRdsVersionMismatched(t *testing.T) {
	tests := []struct {
		testDescription string
		tfOutput        string
		expectedRes     *environment.RdsVersionResults
		wantErr         error
	}{
		{
			"GIVEN an output WITH a version mismatch THEN return the correct strings", mismatchOutput, &environment.RdsVersionResults{
				Versions:               [][]string{{"14.12", "14.10"}},
				ModuleNames:            [][]string{{"rds_example_module_name"}},
				TotalVersionMismatches: 1,
			},
			nil,
		},
		{"GIVEN an output WITHOUT a version mismatch THEN return nil", nomatchOutput, nil, errors.New("terraform is failing but it doesn't look like a rds version mismatch")},
		{"GIVEN an output WITH a version mismatch BUT the mismatch would cause a db UPGRADE THEN return nil", isVersionUpgradeMismatch, nil, errors.New("terraform is failing, but it isn't trying to downgrade the RDS versions so it needs more investigation")},
		{
			"GIVEN an output WITH MULTIPLE version mismatches THEN return the correct string slices", multipleMismatchOutput, &environment.RdsVersionResults{
				Versions:               [][]string{{"14.12", "14.10"}, {"16.18", "16.16"}},
				ModuleNames:            [][]string{{"rds_example_module_name"}, {"rds_example_module_name_2"}},
				TotalVersionMismatches: 2,
			},
			nil,
		},
		{"GIVEN an output with an inconsistent number of versions and modules THEN return nil and an error", inconsistentOutput, nil, errors.New("error: there is an inconistent number of versions vs module names, there should be an even amount but we have 1 sets of versions and 3 module names")},
	}

	for _, tt := range tests {
		t.Run(tt.testDescription, func(t *testing.T) {
			fmt.Printf("Running test: %s\n", tt.testDescription)

			res, err := environment.IsRdsVersionMismatched(tt.tfOutput)
			if tt.wantErr != nil {
				if tt.wantErr.Error() != err.Error() {
					t.Errorf("Expected an error = %v, but did not the receive expected error, actual error = %v", tt.wantErr.Error(), err.Error())
				}
			}

			if tt.expectedRes == nil {
				if res != nil {
					t.Errorf("Expected result to be nil but received %v", res)
				}
				return
			}

			if !reflect.DeepEqual(*tt.expectedRes, *res) {
				t.Errorf("IsRdsVersionMismatched() = %v, want %v", res, tt.expectedRes)
			}
		})
	}
}
