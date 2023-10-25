package deleteutils

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/terraform"
)

type MockTFDAL struct {
	client        *terraform.TerraformCLI
	mockBehaviour string
}

func (r MockTFDAL) Init(ctx context.Context, w io.Writer) error {
	if r.mockBehaviour == "error" {
		return errors.New("big tf init error")
	}
	return nil
}

func (r MockTFDAL) Plan(ctx context.Context, w io.Writer) (bool, error) {
	if r.mockBehaviour == "error" {
		return false, errors.New("big tf plan error")
	}
	return false, nil
}

func (r MockTFDAL) Destroy(ctx context.Context, w io.Writer) error {
	if r.mockBehaviour == "error" {
		return errors.New("big tf destroy error")
	}
	return nil
}

func (r MockTFDAL) WorkspaceDelete(ctx context.Context, workspace string) error {
	if r.mockBehaviour == "error" {
		return errors.New("big tf workspace delete error")
	}
	return nil
}

func Test_initTfCLI(t *testing.T) {
	type args struct {
		tf     *terraform.TerraformCLIConfig
		dryRun bool
	}

	actualDryConfig := &terraform.TerraformCLIConfig{
		ExecPath:   "path/to/tf",
		WorkingDir: "./",
		Workspace:  "default",
		Version:    "0.14.8",
	}

	expectedDryConfig := &terraform.TerraformCLIConfig{
		ExecPath:   "path/to/tf",
		WorkingDir: "./",
		Workspace:  "default",
		Version:    "0.14.8",
		PlanVars:   []tfexec.PlanOption{tfexec.Destroy(true)},
	}

	expectedDestroyConfig := &terraform.TerraformCLIConfig{
		ExecPath:   "path/to/tf",
		WorkingDir: "./",
		Workspace:  "default",
		Version:    "0.14.8",
	}

	actualDestroyConfig := &terraform.TerraformCLIConfig{
		ExecPath:   "path/to/tf",
		WorkingDir: "./",
		Workspace:  "default",
		Version:    "0.14.8",
	}

	actualErrConfig := &terraform.TerraformCLIConfig{
		Version: "",
	}

	expectedDryDestroyCli, _ := terraform.NewTerraformCLI(expectedDryConfig)
	expectedDestroyCli, _ := terraform.NewTerraformCLI(expectedDestroyConfig)

	tests := []struct {
		name    string
		args    args
		want    *terraform.TerraformCLI
		wantErr bool
	}{
		{
			name: "GIVEN terraform config AND dry run is true THEN return terraform cli with plan destroy set",
			args: args{
				tf:     actualDryConfig,
				dryRun: true,
			},
			want:    expectedDryDestroyCli,
			wantErr: false,
		},
		{
			name: "GIVEN terraform config AND dry run is false THEN return terraform cli without plan destroy set",
			args: args{
				tf:     actualDestroyConfig,
				dryRun: false,
			},
			want:    expectedDestroyCli,
			wantErr: false,
		},
		{
			name: "GIVEN terraform config AND there is an error creating the cli THEN return the error",
			args: args{
				tf:     actualErrConfig,
				dryRun: false,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InitTfCLI(tt.args.tf, tt.args.dryRun)

			if (err != nil) != tt.wantErr {
				t.Errorf("initTfCLI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want == nil && !cmp.Equal(got, tt.want) {
				return
			}

			gotTf := got.(*terraform.TerraformCLI).Tf
			gotPlan := got.(*terraform.TerraformCLI).PlanVars

			marshalledGot, _ := json.Marshal(got)
			marshalledWant, _ := json.Marshal(tt.want)
			marshGotPlan, _ := json.Marshal(gotPlan)
			marshWantPlan, _ := json.Marshal(tt.want.PlanVars)
			marshGotTf, _ := json.Marshal(gotTf)
			marshWantTF, _ := json.Marshal(tt.want.Tf)

			if !cmp.Equal(marshalledWant, marshalledGot) && tt.want != nil && !cmp.Equal(marshGotTf, marshWantTF) && !cmp.Equal(marshGotPlan, marshWantPlan) {
				t.Errorf("initTfCLI() = %v, want %v", got, tt.want)
				return
			}
		})
	}
}

func Test_terraformInit(t *testing.T) {
	type args struct {
		terraform     TFDataAccessLayer
		workingDir    string
		mockBehaviour string
	}

	mockClientConfig := &terraform.TerraformCLIConfig{
		ExecPath:   "terraform",
		WorkingDir: "./",
		Workspace:  "default",
		Version:    "0.14.8",
	}

	mockCli, _ := terraform.NewTerraformCLI(mockClientConfig)
	mockClient := &MockTFDAL{client: mockCli}

	errArgs := args{
		terraform:     mockClient,
		workingDir:    "./",
		mockBehaviour: "error",
	}

	validArgs := args{
		terraform:     mockClient,
		workingDir:    "./",
		mockBehaviour: "normal",
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "GIVEN terraform cli AND it errors WHEN running init THEN return the error", args: errArgs, wantErr: true},
		{name: "GIVEN terraform cli AND it errors WHEN running init THEN return the error", args: validArgs, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.mockBehaviour = tt.args.mockBehaviour
			if err := terraformInit(tt.args.terraform, tt.args.workingDir); (err != nil) != tt.wantErr {
				t.Errorf("terraformInit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_terraformDestroy(t *testing.T) {
	type args struct {
		terraform     TFDataAccessLayer
		dryRun        bool
		mockBehaviour string
	}

	mockClientConfig := &terraform.TerraformCLIConfig{
		ExecPath:   "terraform",
		WorkingDir: "./",
		Workspace:  "default",
		Version:    "0.14.8",
	}

	mockCli, _ := terraform.NewTerraformCLI(mockClientConfig)
	mockClient := &MockTFDAL{client: mockCli}

	errDryArgs := args{
		terraform:     mockClient,
		dryRun:        false,
		mockBehaviour: "error",
	}

	validDryArgs := args{
		terraform:     mockClient,
		dryRun:        false,
		mockBehaviour: "normal",
	}

	errArgs := args{
		terraform:     mockClient,
		dryRun:        true,
		mockBehaviour: "error",
	}

	validArgs := args{
		terraform:     mockClient,
		dryRun:        true,
		mockBehaviour: "normal",
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "GIVEN terraform destroy AND it is NOT a dry run THEN run the destroy plan (dry-run) cmd successfully", args: errDryArgs, wantErr: true},
		{name: "GIVEN terraform cli AND it errors WHEN running init THEN return the error", args: validDryArgs, wantErr: false},
		{name: "GIVEN terraform destroy AND it is NOT a dry run WHEN running destroy THEN return the error", args: errArgs, wantErr: true},
		{name: "GIVEN terraform destroy AND it is NOT a dry run THEN run the destroy cmd successfully", args: validArgs, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.mockBehaviour = tt.args.mockBehaviour
			err := terraformDestroy(tt.args.terraform, tt.args.dryRun)
			if (err != nil) != tt.wantErr {
				t.Errorf("terraformDestroy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
