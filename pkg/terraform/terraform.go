// Package terraform implements methods and functions for running
// Terraform commands, such as terraform init/plan/apply.
package terraform

import (
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"

	"github.com/hashicorp/go-version"
	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/fs"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hc-install/src"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
)

var (
	wsFailedToSelectRegexp = regexp.MustCompile(`Failed to select workspace`)
	wsDoesNotExistRegexp   = regexp.MustCompile(`workspace ".*" does not exist`)
	wsAlreadyExists        = regexp.MustCompile(`Workspace ".*" already exists`)
)

// TerraformCLI is the client that wraps around terraform-exec
// to execute Terraform cli commands
type TerraformCLI struct {
	Tf          terraformExec
	WorkingDir  string
	Workspace   string
	ApplyVars   []tfexec.ApplyOption
	DestroyVars []tfexec.DestroyOption
	PlanVars    []tfexec.PlanOption
	InitVars    []tfexec.InitOption
	Redacted    bool
}

// TerraformCLIConfig configures the Terraform client
type TerraformCLIConfig struct {
	// ExecPath is the path to the Terraform executable.
	ExecPath string
	// WorkingDir is the path Terraform will execute in.
	WorkingDir string
	// Worspace is the Terraform workspace to use.
	Workspace string
	// ApplyVars allows you to group apply options passed to Terraform.
	ApplyVars []tfexec.ApplyOption
	// DestroyVars allows you to group destroy options passed to Terraform.
	DestroyVars []tfexec.DestroyOption
	// PlanVars allows you to group plan variables passed to Terraform.
	PlanVars []tfexec.PlanOption
	// InitVars allows you to group init variables passed to Terraform.
	InitVars []tfexec.InitOption
	// Version is the version of Terraform to use.
	Version string
	// Redacted is the flag to enable/disable redacting the terraform before printing output.
	Redacted bool
}

// NewTerraformCLI creates a terraform-exec client and configures and
// initializes a new Terraform client
func NewTerraformCLI(config *TerraformCLIConfig) (*TerraformCLI, error) {
	if config == nil {
		return nil, errors.New("TerraformCLIConfig cannot be nil - no meaningful default values")
	}

	if config.Version == "" {
		return nil, errors.New("version cannot be empty")
	}

	if config.ExecPath == "" {
		i := install.NewInstaller()
		v := version.Must(version.NewVersion(config.Version))

		execPath, err := i.Ensure(context.TODO(), []src.Source{
			&fs.ExactVersion{
				Product: product.Terraform,
				Version: v,
			},
			&releases.ExactVersion{
				Product: product.Terraform,
				Version: v,
			},
		})
		if err != nil {
			return nil, err
		}

		config.ExecPath = execPath

		defer func() {
			if err := i.Remove(context.TODO()); err != nil {
				return
			}
		}()
	}

	tf, err := tfexec.NewTerraform(config.WorkingDir, config.ExecPath)
	if err != nil {
		return nil, err
	}

	client := &TerraformCLI{
		Tf:          tf,
		WorkingDir:  config.WorkingDir,
		Workspace:   config.Workspace,
		ApplyVars:   config.ApplyVars,
		DestroyVars: config.DestroyVars,
		PlanVars:    config.PlanVars,
		InitVars:    config.InitVars,
		Redacted:    config.Redacted,
	}

	return client, nil
}

// Init initializes by executing the cli command `terraform init` and
// `terraform workspace new <name>`
func (t *TerraformCLI) Init(ctx context.Context, w io.Writer) error {
	var wsCreated bool

	t.Tf.SetStdout(w)
	t.Tf.SetStderr(w)
	// This is special handling for when the workspace has been detected in
	// .terraform/environment with a non-existing state. This case is common
	// when the state for the workspace has been deleted.
	// https://github.com/hashicorp/terraform/issues/21393
TF_INIT_AGAIN:
	if err := t.Tf.Init(ctx); err != nil {
		matchedFailedToSelect := wsFailedToSelectRegexp.MatchString(err.Error())
		matchedDoesNotExist := wsDoesNotExistRegexp.MatchString(err.Error())
		if matchedFailedToSelect || matchedDoesNotExist {
			fmt.Println("workspace was detected without state, " +
				"creating new workspace and attempting Terraform init again")
			if err := t.Tf.WorkspaceNew(ctx, t.Workspace); err != nil {
				return err
			}

			if !wsCreated {
				wsCreated = true
				goto TF_INIT_AGAIN
			}
		}
		return err
	}

	if !wsCreated {
		if err := t.Tf.WorkspaceNew(ctx, t.Workspace); err != nil {
			matchedAlreadyExists := wsAlreadyExists.MatchString(err.Error())
			if err != nil && !matchedAlreadyExists {
				return err
			}
		}
	}

	if err := t.Tf.WorkspaceSelect(ctx, t.Workspace); err != nil {
		return err
	}

	return nil
}

// Apply executes the cli command `terraform apply` for a given workspace
func (t *TerraformCLI) Apply(ctx context.Context, w io.Writer) error {
	t.Tf.SetStdout(w)
	t.Tf.SetStderr(w)

	if err := t.Tf.Apply(ctx, t.ApplyVars...); err != nil {
		return err
	}

	return nil
}

// Destroy executes the cli command `terraform destroy` for a given workspace
func (t *TerraformCLI) Destroy(ctx context.Context, w io.Writer) error {
	t.Tf.SetStdout(w)
	t.Tf.SetStderr(w)

	if err := t.Tf.Destroy(ctx, t.DestroyVars...); err != nil {
		return fmt.Errorf("failed to destroy component in %s: %w", t.Workspace, err)
	}

	return nil
}

// Plan executes the cli command `terraform plan` for a given workspace
func (t *TerraformCLI) Plan(ctx context.Context, w io.Writer) (bool, error) {
	t.Tf.SetStdout(w)
	t.Tf.SetStderr(w)

	diff, err := t.Tf.Plan(ctx, t.PlanVars...)
	if err != nil {
		return false, err
	}

	return diff, nil
}

// Plan executes the cli command `terraform plan` for a given workspace
func (t *TerraformCLI) Output(ctx context.Context, w io.Writer) (map[string]tfexec.OutputMeta, error) {
	// Often times, the output is not needed, so the caller can specify a null writer to ignore.
	t.Tf.SetStdout(w)
	return t.Tf.Output(ctx)
}

// Show reads the default state path and outputs the state
func (t *TerraformCLI) Show(ctx context.Context, w io.Writer) (*tfjson.State, error) {
	return t.Tf.Show(ctx)
}

// StateList loop over the state and builds a state list
func (t *TerraformCLI) StateList(state *tfjson.State) []string {
	stateList := []string{}

	for _, resource := range state.Values.RootModule.Resources {
		stateList = append(stateList, resource.Address)
	}
	for _, childResources := range state.Values.RootModule.ChildModules {
		for _, modules := range childResources.Resources {
			stateList = append(stateList, modules.Address)
		}
	}
	return stateList
}

func (t *TerraformCLI) WorkspaceDelete(ctx context.Context, workspace string) error {
	if err := t.Tf.WorkspaceSelect(ctx, "default"); err != nil {
		return err
	}

	if err := t.Tf.WorkspaceDelete(ctx, workspace); err != nil {
		return err
	}

	return nil
}
