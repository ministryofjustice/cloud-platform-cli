// Package terraform implements methods and functions for running
// Terraform commands, such as terraform init/plan/apply.
//
// The intention of this package is to call and run inside a CI/CD
// pipeline.
package terraform

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	mocks "github.com/ministryofjustice/cloud-platform-cli/pkg/mocks/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func NewTestTerraformCLI(config *TerraformCLIConfig, tfMock *mocks.TerraformExec) *TerraformCLI {
	if tfMock == nil {
		m := new(mocks.TerraformExec)
		m.On("SetStdout", mock.Anything).Once()
		m.On("SetStderr", mock.Anything).Once()
		m.On("Init", mock.Anything).Return(nil)
		m.On("Apply", mock.Anything).Return(nil)
		m.On("Destroy", mock.Anything).Return(nil)
		m.On("Plan", mock.Anything).Return(true, nil)
		m.On("Output", mock.Anything).Return(nil, nil)
		m.On("Show", mock.Anything).Return(nil, nil)
		m.On("WorkspaceNew", mock.Anything, mock.Anything).Return(nil)
		m.On("WorkspaceDelete", mock.Anything, mock.Anything).Return(nil)
		tfMock = m
	}

	tfCli := &TerraformCLI{
		Tf:         tfMock,
		WorkingDir: "test/working/dir",
		Workspace:  "test-workspace",
	}

	if config == nil {
		return tfCli
	}

	if config.WorkingDir != "" {
		tfCli.WorkingDir = config.WorkingDir
	}
	if config.Workspace != "" {
		tfCli.Workspace = config.Workspace
	}

	return tfCli
}

func TestNewTerraformCLI(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		expectError bool
		config      *TerraformCLIConfig
	}{
		{
			"error nil config",
			true,
			nil,
		},
		{
			"version not set",
			true,
			&TerraformCLIConfig{
				WorkingDir: "",
				Workspace:  "default",
				Version:    "",
			},
		},
		{
			"terraform-exec error: no working dir",
			true,
			&TerraformCLIConfig{
				ExecPath:   "path/to/tf",
				WorkingDir: "",
				Workspace:  "default",
				Version:    "0.14.8",
			},
		},
		{
			"happy path",
			false,
			&TerraformCLIConfig{
				ExecPath:   "path/to/tf",
				WorkingDir: "./",
				Workspace:  "my-workspace",
				Version:    "0.14.8",
			},
		},
		{
			"null execPath path",
			false,
			&TerraformCLIConfig{
				ExecPath:   "",
				WorkingDir: "./",
				Workspace:  "my-workspace",
				Version:    "0.14.8",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := NewTerraformCLI(tc.config)

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, actual)
		})
	}
}

func TestTerraformCLI_Init(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		expectError bool
		config      *TerraformCLIConfig
		initErr     error
		wsErr       error
	}{
		{
			"happy path",
			false,
			&TerraformCLIConfig{},
			nil,
			nil,
		},
		{
			"init err",
			true,
			&TerraformCLIConfig{},
			errors.New("init error"),
			nil,
		},
		{
			"workspace-new error: unknown error",
			true,
			&TerraformCLIConfig{},
			nil,
			errors.New("workspace-new error"),
		},
		{
			"workspace-new: already exists",
			false,
			&TerraformCLIConfig{},
			nil,
			errors.New(`Workspace "foobar" already exists`),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mocks.TerraformExec)
			m.On("SetStdout", mock.Anything).Once()
			m.On("SetStderr", mock.Anything).Once()
			m.On("Init", mock.Anything).Return(tc.initErr).Once()
			m.On("WorkspaceNew", mock.Anything, mock.Anything).Return(tc.wsErr)
			m.On("WorkspaceSelect", mock.Anything, mock.Anything).Return(nil)

			tfCli := NewTestTerraformCLI(tc.config, m)
			ctx := context.Background()
			var out bytes.Buffer
			err := tfCli.Init(ctx, &out)

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			m.AssertExpectations(t)
		})
	}
}

func TestTerraformCLIInit_HandleWorkspaceError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		initErr error
	}{
		{
			"workspace failed to select",
			errors.New(`Initializing the backend...

The currently selected workspace (test-workspace) does not exist.
This is expected behavior when the selected workspace did not have an
existing non-empty state. Please enter a number to select a workspace:

1. default

Enter a value:

Error: Failed to select workspace: input not a valid number`),
		},
		{
			"workspace does not exist",
			errors.New(`exit status 1

Error: Currently selected workspace "some-task" does not exist


`),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mocks.TerraformExec)
			var initCount int
			m.On("SetStdout", mock.Anything).Once()
			m.On("SetStderr", mock.Anything).Once()
			m.On("Init", mock.Anything).Return(func(context.Context, ...tfexec.InitOption) error {
				initCount++
				if initCount == 1 {
					return tc.initErr
				}
				return nil
			}).Twice()
			m.On("WorkspaceNew", mock.Anything, mock.Anything).Return(nil)
			m.On("WorkspaceSelect", mock.Anything, mock.Anything).Return(nil)

			tfCli := NewTestTerraformCLI(&TerraformCLIConfig{}, m)
			ctx := context.Background()
			var out bytes.Buffer
			err := tfCli.Init(ctx, &out)
			assert.NoError(t, err)
			m.AssertExpectations(t)
		})
	}
}

func TestTerraformCLI_Apply(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		expectError bool
		config      *TerraformCLIConfig
	}{
		{
			"happy path",
			false,
			&TerraformCLIConfig{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mocks.TerraformExec)
			m.On("SetStdout", mock.Anything).Once()
			m.On("SetStderr", mock.Anything).Once()
			tfCli := NewTestTerraformCLI(tc.config, nil)
			ctx := context.Background()
			var out bytes.Buffer
			err := tfCli.Apply(ctx, &out)

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestTerraformCLI_Destroy(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		expectError bool
		config      *TerraformCLIConfig
	}{
		{
			"happy path",
			false,
			&TerraformCLIConfig{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mocks.TerraformExec)
			m.On("SetStdout", mock.Anything).Once()
			m.On("SetStderr", mock.Anything).Once()
			tfCli := NewTestTerraformCLI(tc.config, nil)
			ctx := context.Background()
			var out bytes.Buffer
			err := tfCli.Destroy(ctx, &out)

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestTerraformCLI_Plan(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		expectError bool
		config      *TerraformCLIConfig
	}{
		{
			"happy path",
			false,
			&TerraformCLIConfig{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tfCli := NewTestTerraformCLI(tc.config, nil)
			m := new(mocks.TerraformExec)
			m.On("SetStdout", mock.Anything).Once()
			m.On("SetStderr", mock.Anything).Once()
			ctx := context.Background()
			var out bytes.Buffer
			_, err := tfCli.Plan(ctx, &out)

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestTerraformCLI_Output(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		expectError bool
		config      *TerraformCLIConfig
	}{
		{
			"happy path",
			false,
			&TerraformCLIConfig{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tfCli := NewTestTerraformCLI(tc.config, nil)
			ctx := context.Background()
			_, err := tfCli.Output(ctx, nil)

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestTerraformCLI_Show(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		expectError bool
		config      *TerraformCLIConfig
	}{
		{
			"happy path",
			false,
			&TerraformCLIConfig{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tfCli := NewTestTerraformCLI(tc.config, nil)
			ctx := context.Background()
			_, err := tfCli.Show(ctx, nil)

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestTerraformCLI_StateList(t *testing.T) {
	type args struct {
		state *tfjson.State
	}
	tests := []struct {
		name   string
		config *TerraformCLIConfig
		args   args
		want   []string
	}{
		{
			name:   "happy path",
			config: &TerraformCLIConfig{},
			args: args{
				state: &tfjson.State{
					Values: &tfjson.StateValues{
						RootModule: &tfjson.StateModule{
							Resources: []*tfjson.StateResource{{
								Address:         "null_resource.foo",
								AttributeValues: map[string]interface{}{},
							}},
						},
					},
				},
			},

			want: []string{"null_resource.foo"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(mocks.TerraformExec)
			tfCli := NewTestTerraformCLI(tt.config, m)
			if got := tfCli.StateList(tt.args.state); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TerraformCLI.StateList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTerraformCLI_WorkspaceDelete(t *testing.T) {
	type args struct {
		ctx       context.Context
		workspace string
		mockError error
	}
	tests := []struct {
		name    string
		config  *TerraformCLIConfig
		args    args
		wantErr bool
	}{
		{
			"GIVEN a terraform workspace THEN delete it successfully", &TerraformCLIConfig{
				WorkingDir: "./",
				Workspace:  "test-workspace",
				Version:    "1.2.5",
			}, args{context.Background(), "test-workspace", nil}, false,
		},
		{
			"GIVEN a terraform workspace AND the delete errors THEN return the error", &TerraformCLIConfig{
				WorkingDir: "./",
				Workspace:  "test-workspace",
				Version:    "1.2.5",
			}, args{context.Background(), "test-workspace", errors.New("deleting tf workspace")}, true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(mocks.TerraformExec)
			m.On("WorkspaceDelete", mock.Anything, mock.Anything).Return(tt.args.mockError)
			m.On("WorkspaceSelect", mock.Anything, mock.Anything).Return(nil)
			tfCli := NewTestTerraformCLI(tt.config, m)
			if err := tfCli.WorkspaceDelete(tt.args.ctx, tt.args.workspace); (err != nil) != tt.wantErr {
				t.Errorf("TerraformCLI.WorkspaceDelete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
