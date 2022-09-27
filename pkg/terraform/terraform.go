// Package terraform implements methods and functions for running
// Terraform commands, such as terraform init/plan/apply.
//
// The intention of this package is to call and run inside a CI/CD
// pipeline.
package terraform

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/hashicorp/go-version"
	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/fs"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hc-install/src"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/client"
	log "github.com/sirupsen/logrus"
)

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

// Options are the options to pass to Terraform plan and apply.
type Options struct {
	// Apply allows you to group apply options passed to Terraform.
	ApplyVars []tfexec.ApplyOption
	// Init allows you to group init options passed to Terraform.
	InitVars []tfexec.InitOption
	// Version is the version of Terraform to use.
	Version string
	// ExecPath is the path to the Terraform executable.
	ExecPath string
	// Workspace is the name of the Terraform workspace to use.
	Workspace string
	// FilePath is the location of the cloud-platform-infrastructure repository.
	// This repository contains all the Terraform used to create the cluster.
	FilePath     string
	DirStructure []string
}

func NewOptions(version, workspace string) (*Options, error) {
	if version == "" {
		return nil, fmt.Errorf("terraform version is required")
	}
	tf := &Options{
		Version:   version,
		Workspace: workspace,
	}

	if err := tf.CreateTerraformObj(); err != nil {
		return nil, err
	}

	return tf, nil
}

// Terraform creates terraform command to be executed
func (s *Commander) Terraform(args ...string) (*CmdOutput, error) {
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	var exitCode int
	var err error

	contextLogger := log.WithFields(log.Fields{
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
		log.Error(output.Stderr)
		return err
	}

	if p {
		log.Info(output.Stdout)
	}

	return nil
}

// SelectWs is used to select certain workspace.
func (s *Commander) SelectWs(ws string) error {
	output, err := s.Terraform("workspace", "select", ws)
	if err != nil {
		return err
	}

	log.Info(output.Stdout)

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

// createTerraformObj creates a Terraform object using the version passed as a string.
func (terraform *Options) CreateTerraformObj() error {
	i := install.NewInstaller()
	v := version.Must(version.NewVersion(terraform.Version))

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
		return err
	}

	defer i.Remove(context.TODO())

	terraform.ExecPath = execPath
	terraform.ApplyVars = []tfexec.ApplyOption{
		tfexec.Parallelism(1),
	}

	return nil
}

func (terraform *Options) initAndApply(tf *tfexec.Terraform, creds *client.AwsCredentials) error {
	fmt.Printf("Init terraform in directory %s\n", tf.WorkingDir())
	err := terraform.initialise(tf)
	if err != nil {
		return fmt.Errorf("failed to init terraform: %w", err)
	}

	fmt.Printf("Apply terraform in directory %s\n", tf.WorkingDir())
	err = terraform.executeApply(tf)
	if err != nil {
		return fmt.Errorf("failed to apply: %w", err)
	}

	return nil
}

func (terraform *Options) output(tf *tfexec.Terraform) (map[string]tfexec.OutputMeta, error) {
	// We don't want terraform to print out the output here as the package doesn't respect the secret flag.
	tf.SetStdout(nil)
	tf.SetStderr(nil)
	output, err := tf.Output(context.TODO())
	if err != nil {
		if strings.Contains(err.Error(), "plugin") || strings.Contains(err.Error(), "init") {
			fmt.Println("Init again, due to failure")
			err = tf.Init(context.TODO(), terraform.InitVars...)
			if err != nil {
				return nil, fmt.Errorf("failed to init: %w", err)
			}
			output, err = tf.Output(context.TODO())
			if err != nil {
				return nil, fmt.Errorf("failed to create output: %w", err)
			}
			return nil, fmt.Errorf("failed to show terraform output: %w", err)
		}
	}
	return output, nil
}

// intialise performs the `terraform init` function.
func (terraform *Options) initialise(tf *tfexec.Terraform) error {
	terraform.InitVars = append(terraform.InitVars, tfexec.Reconfigure(true))
	err := tf.Init(context.TODO(), terraform.InitVars...)
	// Handle no plugin error
	if err != nil {
		if err := tf.Init(context.TODO(), terraform.InitVars...); err != nil {
			return fmt.Errorf("failed to init: %w", err)
		}
	}

	return terraform.setWorkspace(tf)
}

func (terraform *Options) setWorkspace(tf *tfexec.Terraform) error {
	list, _, err := tf.WorkspaceList(context.TODO())
	if err != nil {
		return err
	}

	for _, ws := range list {
		if ws == terraform.Workspace {
			err = tf.WorkspaceSelect(context.TODO(), terraform.Workspace)
			if err != nil {
				return err
			}
			return nil
		}
	}

	err = tf.WorkspaceNew(context.TODO(), terraform.Workspace)
	if err != nil {
		return err
	}

	return nil
}

func (terraform *Options) executeApply(tf *tfexec.Terraform) error {
	// TODO: Pass the argumnet via the cluster package
	// if strings.Contains(tf.WorkingDir(), "eks") && fast {
	// 	terraform.ApplyVars = append(terraform.ApplyVars, tfexec.Var(fmt.Sprintf("%s=%v", "enable_oidc_associate", false)))
	// }

	err := tf.Apply(context.TODO(), terraform.ApplyVars...)
	// handle a case where you need to init again
	if err != nil {
		fmt.Println("Init again, due to failure")
		err = tf.Init(context.TODO(), terraform.InitVars...)
		if err != nil {
			return fmt.Errorf("failed to init: %w", err)
		}
		err = tf.Apply(context.TODO(), terraform.ApplyVars...)
		if err != nil {
			return fmt.Errorf("failed to apply: %w", err)
		}
	}

	return nil
}

func (terraform *Options) Apply(exec *tfexec.Terraform, creds *client.AwsCredentials) (map[string]tfexec.OutputMeta, error) {
	// Write the output to the terminal.
	exec.SetStdout(os.Stdout)
	exec.SetStdout(os.Stderr)

	err := terraform.initAndApply(exec, creds)
	if err != nil {
		return nil, fmt.Errorf("an error occurred attempting to init and apply %w", err)
	}

	output, err := terraform.output(exec)
	if err != nil {
		return nil, fmt.Errorf("failed to get output: %w", err)
	}

	return output, nil
}
