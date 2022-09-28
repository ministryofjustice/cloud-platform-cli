package terraform

import (
	"testing"

	"github.com/hashicorp/terraform-exec/tfexec"
)

func TestNewOptions(t *testing.T) {
	options, err := NewOptions("1.1.1", "testWorkspace")
	if err != nil {
		t.Errorf("newTerraformOptions() error = %v", "expected error")
	}

	if options.Version != "1.1.1" {
		t.Errorf("newTerraformOptions() error = %v", "terraform version not set")
	}

	_, err = NewOptions("", "testWorkspace")
	if err == nil {
		t.Errorf("was expecting an error. newTerraformOptions() error = %v", "expected error")
	}
}

func TestOptions_CreateTerraformObj(t *testing.T) {
	type fields struct {
		ApplyVars []tfexec.ApplyOption
		InitVars  []tfexec.InitOption
		Version   string
		ExecPath  string
		Workspace string
		FilePath  string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "test create terraform object",
			fields: fields{
				Version:   "1.1.1",
				Workspace: "testWorkspace",
			},
			wantErr: false,
		},
		{
			name: "test create terraform object with invalid version",
			fields: fields{
				Version:   "0.0.1",
				Workspace: "testWorkspace",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			terraform := &Options{
				ApplyVars: tt.fields.ApplyVars,
				InitVars:  tt.fields.InitVars,
				Version:   tt.fields.Version,
				ExecPath:  tt.fields.ExecPath,
				Workspace: tt.fields.Workspace,
				FilePath:  tt.fields.FilePath,
			}
			if err := terraform.CreateTerraformObj(); (err != nil) != tt.wantErr {
				t.Errorf("Options.CreateTerraformObj() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
