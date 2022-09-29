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
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	"github.com/hashicorp/go-version"
	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/fs"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hc-install/src"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/sirupsen/logrus"
)

var (
	wsFailedToSelectRegexp = regexp.MustCompile(`Failed to select workspace`)
	wsDoesNotExistRegexp   = regexp.MustCompile(`workspace ".*" does not exist`)
	pluginNotFound         = regexp.MustCompile(`Unrecognized remote plugin message`)
)

// TerraformCLI is the client that wraps around terraform-exec
// to execute Terraform cli commands
type TerraformCLI struct {
	tf         terraformExec
	workingDir string
	workspace  string
	// Apply allows you to group apply options passed to Terraform.
	applyVars []tfexec.ApplyOption
	// Init allows you to group init options passed to Terraform.
	initVars []tfexec.InitOption
}

// TerraformCLIConfig configures the Terraform client
type TerraformCLIConfig struct {
	ExecPath   string
	WorkingDir string
	Workspace  string
	// Apply allows you to group apply options passed to Terraform.
	ApplyVars []tfexec.ApplyOption
	// Init allows you to group init options passed to Terraform.
	InitVars []tfexec.InitOption
	// Version is the version of Terraform to use.
	Version string
	// ExecPath is the path to the Terraform executable.
}

// Commander struct holds all data required to execute terraform.
type Commander struct {
	cmdDir          string
	cmdEnv          []string
	AccessKeyID     string
	SecretAccessKey string
	Workspace       string
	VarFile         string
	DisplayTfOutput bool
	BulkTfPaths     string
}

// NewTerraformCLI creates a terraform-exec client and configures and
// initializes a new Terraform client
func NewTerraformCLI(config *TerraformCLIConfig) (*TerraformCLI, error) {
	if config == nil {
		return nil, errors.New("TerraformCLIConfig cannot be nil - no meaningful default values")
	}

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

	defer func() {
		if err := i.Remove(context.TODO()); err != nil {
			return
		}
	}()

	config.ExecPath = execPath

	tf, err := tfexec.NewTerraform(config.WorkingDir, config.ExecPath)
	if err != nil {
		return nil, err
	}

	tf.SetStdout(log.Writer())
	tf.SetStderr(log.Writer())

	client := &TerraformCLI{
		tf:         tf,
		workingDir: config.WorkingDir,
		workspace:  config.Workspace,
		applyVars:  config.ApplyVars,
		initVars:   config.InitVars,
	}

	return client, nil
}

// Init initializes by executing the cli command `terraform init` and
// `terraform workspace new <name>`
func (t *TerraformCLI) Init(ctx context.Context) error {
	var wsCreated bool

	// This is special handling for when the workspace has been detected in
	// .terraform/environment with a non-existing state. This case is common
	// when the state for the workspace has been deleted.
	// https://github.com/hashicorp/terraform/issues/21393
TF_INIT_AGAIN:
	if err := t.tf.Init(ctx); err != nil {
		var wsErr *tfexec.ErrNoWorkspace
		matchedFailedToSelect := wsFailedToSelectRegexp.MatchString(err.Error())
		matchedDoesNotExist := wsDoesNotExistRegexp.MatchString(err.Error())
		pluginErr := pluginNotFound.MatchString(err.Error())
		if matchedFailedToSelect || matchedDoesNotExist || pluginErr || errors.As(err, &wsErr) {
			if err := t.tf.WorkspaceNew(ctx, t.workspace); err != nil {
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
		err := t.tf.WorkspaceNew(ctx, t.workspace)
		if err != nil {
			var wsErr *tfexec.ErrWorkspaceExists
			if !errors.As(err, &wsErr) {
				return err
			}
		}
	}

	if err := t.tf.WorkspaceSelect(ctx, t.workspace); err != nil {
		return err
	}

	return nil
}

// Apply executes the cli command `terraform apply` for a given workspace
func (t *TerraformCLI) Apply(ctx context.Context) error {
	return t.tf.Apply(ctx, t.applyVars...)
}

// Plan executes the cli command `terraform plan` for a given workspace
func (t *TerraformCLI) Plan(ctx context.Context) (bool, error) {
	return t.tf.Plan(ctx)
}

// Plan executes the cli command `terraform plan` for a given workspace
func (t *TerraformCLI) Output(ctx context.Context) (map[string]tfexec.OutputMeta, error) {
	return t.tf.Output(ctx)
}

// Terraform creates terraform command to be executed
func (s *Commander) Terraform(args ...string) (*CmdOutput, error) {
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	var exitCode int
	var err error

	contextLogger := logrus.WithFields(logrus.Fields{
		"err":    err,
		"stdout": stdoutBuf.String(),
		"stderr": stderrBuf.String(),
		"dir":    s.cmdDir,
	})

	cmd := exec.Command("terraform", args...)

	if s.cmdDir != "" {
		cmd.Dir = s.cmdDir
	}

	if s.cmdEnv != nil {
		cmd.Env = s.cmdEnv
	}

	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
		}
		contextLogger.Error("cmd.Run() failed with:")
		contextLogger.Error(cmd.Stderr)

		cmdOutput := CmdOutput{
			Stdout:   stdoutBuf.String(),
			Stderr:   stderrBuf.String(),
			ExitCode: exitCode,
		}
		return &cmdOutput, err
	} else {
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
	}

	cmdOutput := CmdOutput{
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
		ExitCode: exitCode,
	}

	if cmdOutput.ExitCode != 0 {
		return &cmdOutput, err
	} else {
		return &cmdOutput, nil
	}
}

// Init executes terraform init.
func (s *Commander) Init(p bool) error {
	output, err := s.Terraform("init")
	if err != nil {
		logrus.Error(output.Stderr)
		return err
	}

	if p {
		logrus.Info(output.Stdout)
	}

	return nil
}

// SelectWs is used to select certain workspace.
func (s *Commander) SelectWs(ws string) error {
	output, err := s.Terraform("workspace", "select", ws)
	if err != nil {
		return err
	}

	logrus.Info(output.Stdout)

	return nil
}

// CheckDivergence check that there are not changes within certain state, if there are
// it will return non-zero and pipeline will fail.
func (s *Commander) CheckDivergence() error {
	err := s.Init(true)
	if err != nil {
		return err
	}

	err = s.SelectWs(s.Workspace)
	if err != nil {
		return err
	}

	var cmd []string

	// Check if user provided a terraform var-file.
	if s.VarFile != "" {
		cmd = []string{fmt.Sprintf("-var-file=%s", s.VarFile)}
	}

	arg := []string{
		"plan",
		"-no-color",
		"-detailed-exitcode",
	}

	arg = append(arg, cmd...)

	output, err := s.Terraform(arg...)
	if err != nil {
		// there is a drift and hence the cmd returns err with exitcode 2
		if output.ExitCode == 2 {
			if s.DisplayTfOutput && output != nil {
				output.redacted()
			}
		}
		return err
	}

	if s.DisplayTfOutput && output != nil {
		output.redacted()
	}

	if output.ExitCode == 0 {
		return nil
	}
	return err
}

// Apply executes terraform apply.
func (s *Commander) Apply() error {
	err := s.Init(true)
	if err != nil {
		return err
	}

	err = s.SelectWs(s.Workspace)
	if err != nil {
		return err
	}

	var cmd []string

	// Check if user provided a terraform var-file
	if s.VarFile != "" {
		cmd = []string{fmt.Sprintf("-var-file=%s", s.VarFile)}
	}

	arg := []string{
		"apply",
		"-no-color",
		"-auto-approve",
	}

	arg = append(arg, cmd...)

	output, err := s.Terraform(arg...)
	if err != nil {
		return err
	}

	if s.DisplayTfOutput && output != nil {
		output.redacted()
	}

	if output.ExitCode == 0 {
		return nil
	}

	return err
}

// Plan executes terraform plan
func (s *Commander) Plan() error {
	err := s.Init(false)
	if err != nil {
		return err
	}

	err = s.SelectWs(s.Workspace)
	if err != nil {
		return err
	}

	var cmd []string

	// Check if user provided a terraform var-file
	if s.VarFile != "" {
		cmd = []string{fmt.Sprintf("-var-file=%s", s.VarFile)}
	}

	arg := []string{
		"plan",
		"-no-color",
	}

	arg = append(arg, cmd...)

	output, err := s.Terraform(arg...)
	if err != nil {
		return err
	}

	if s.DisplayTfOutput && output != nil {
		output.redacted()
	}

	if output.ExitCode == 0 {
		return nil
	}

	return err
}

// Workspaces return the workspaces within the state.
func (c *Commander) workspaces() ([]string, error) {
	arg := []string{
		"workspace",
		"list",
	}

	output, err := c.Terraform(arg...)
	if err != nil {
		return nil, err
	}

	ws := strings.Split(output.Stdout, " ")
	for i := range ws {
		ws[i] = strings.TrimSpace(ws[i])
	}

	return ws, nil
}
