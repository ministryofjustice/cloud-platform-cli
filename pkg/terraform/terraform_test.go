package terraform

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-exec/tfexec"
	mocks "github.com/ministryofjustice/cloud-platform-cli/pkg/mocks/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func NewTestTerraformCLI(config *TerraformCLIConfig, tfMock *mocks.TerraformExec) *TerraformCLI {
	if tfMock == nil {
		m := new(mocks.TerraformExec)
		m.On("Init", mock.Anything).Return(nil)
		m.On("Apply", mock.Anything).Return(nil)
		m.On("Plan", mock.Anything).Return(true, nil)
		m.On("Output", mock.Anything).Return(nil, nil)
		m.On("WorkspaceNew", mock.Anything, mock.Anything).Return(nil)
		tfMock = m
	}

	tfCli := &TerraformCLI{
		tf:         tfMock,
		workingDir: "test/working/dir",
		workspace:  "test-workspace",
	}

	if config == nil {
		return tfCli
	}

	if config.WorkingDir != "" {
		tfCli.workingDir = config.WorkingDir
	}
	if config.Workspace != "" {
		tfCli.workspace = config.Workspace
	}

	return tfCli
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
			&tfexec.ErrWorkspaceExists{Name: "workspace-name"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mocks.TerraformExec)
			m.On("Init", mock.Anything).Return(tc.initErr).Once()
			m.On("WorkspaceNew", mock.Anything, mock.Anything).Return(tc.wsErr)
			m.On("WorkspaceSelect", mock.Anything, mock.Anything).Return(nil)

			tfCli := NewTestTerraformCLI(tc.config, m)
			ctx := context.Background()
			err := tfCli.Init(ctx)

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
			err := tfCli.Init(ctx)
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
			tfCli := NewTestTerraformCLI(tc.config, nil)
			ctx := context.Background()
			err := tfCli.Apply(ctx)

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
			ctx := context.Background()
			_, err := tfCli.Plan(ctx)

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
			_, err := tfCli.Output(ctx)

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}