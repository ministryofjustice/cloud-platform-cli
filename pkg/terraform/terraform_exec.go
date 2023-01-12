package terraform

import (
	"context"
	"io"

	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
)

//go:generate mockery --name=terraformExec  --structname=TerraformExec --output=pkg/mocks/terraform --dir pkg/terraform

var _ terraformExec = (*tfexec.Terraform)(nil)

// terraformExec describes the interface for terraform-exec, the SDK for
// Terraform CLI: https://github.com/hashicorp/terraform-exec
type terraformExec interface {
	SetStdout(w io.Writer)
	SetStderr(w io.Writer)
	Init(ctx context.Context, opts ...tfexec.InitOption) error
	Apply(ctx context.Context, opts ...tfexec.ApplyOption) error
	Plan(ctx context.Context, opts ...tfexec.PlanOption) (bool, error)
	Output(ctx context.Context, opts ...tfexec.OutputOption) (map[string]tfexec.OutputMeta, error)
	WorkspaceNew(ctx context.Context, workspace string, opts ...tfexec.WorkspaceNewCmdOption) error
	WorkspaceSelect(ctx context.Context, workspace string) error
	Show(ctx context.Context, opts ...tfexec.ShowOption) (*tfjson.State, error)
}
