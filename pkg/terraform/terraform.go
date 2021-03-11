// Package terraform implements methods and functions for running
// Terraform commands, such as terraform init/plan/apply.
//
// The intention of this package is to call and run inside a CI/CD
// pipeline.
package terraform

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

// Commander struct holds all data required to execute terraform.
type Commander struct {
	action          string
	cmd             []string
	cmdDir          string
	cmdEnv          []string
	AccessKeyID     string
	SecretAccessKey string
	Workspace       string
	VarFile         string
	DisplayTfOutput bool
	BulkTfPaths     string
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
		cmd = append([]string{fmt.Sprintf("-var-file=%s", s.VarFile)})
	}

	arg := append(
		[]string{
			"plan",
			"-no-color",
			"-detailed-exitcode",
		},
		cmd...,
	)

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
		cmd = append([]string{fmt.Sprintf("-var-file=%s", s.VarFile)})
	}

	arg := append(
		[]string{
			"apply",
			"-no-color",
			"-auto-approve",
		},
		cmd...,
	)

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
		cmd = append([]string{fmt.Sprintf("-var-file=%s", s.VarFile)})
	}

	arg := append(
		[]string{
			"plan",
			"-no-color",
		},
		cmd...,
	)

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

	ws := strings.Split(output.Stdout, "\n")
	for i := range ws {
		ws[i] = strings.TrimSpace(ws[i])
	}

	return ws, nil
}

// BulkPlan executes plan against all directories that changed in the PR.
func (c *Commander) BulkPlan() error {
	dirs, err := targetDirs(c.BulkTfPaths)
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		fmt.Printf("\n")
		fmt.Println("#########################################################################")
		fmt.Printf("PLAN FOR DIRECTORY: %v\n", dir)
		fmt.Println("#########################################################################")
		fmt.Printf("\n")
		c.cmdDir = dir
		err := c.Init(false)
		if err != nil {
			return err
		}

		ws, err := c.workspaces()
		if err != nil {
			return err
		}

		if contains(ws, "live-1") {
			log.WithFields(log.Fields{"dir": dir}).Info("Using live-1 context with: KUBE_CTX=live-1.cloud-platform.service.justice.gov.uk")

			c.cmdEnv = append(os.Environ(), "KUBE_CTX=live-1.cloud-platform.service.justice.gov.uk")
			c.Workspace = "live-1"

			err := c.Plan()
			if err != nil {
				return err
			}
		} else if contains(ws, "manager") {
			log.WithFields(log.Fields{"dir": dir}).Info("Using manager context with: KUBE_CTX=manager.cloud-platform.service.justice.gov.uk")

			c.cmdEnv = append(os.Environ(), "KUBE_CTX=manager.cloud-platform.service.justice.gov.uk")
			c.Workspace = "manager"

			err := c.Plan()
			if err != nil {
				return err
			}
		} else {
			log.WithFields(log.Fields{"dir": dir}).Info("No context, normal terraform plan")
			err := c.Plan()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// BulkApply executes teraform apply against all directories that changed in the PR.
func (c *Commander) BulkApply() error {
	dirs, err := targetDirs(c.BulkTfPaths)
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		fmt.Printf("\n")
		fmt.Println("#########################################################################")
		fmt.Printf("APPLY FOR DIRECTORY: %v\n", dir)
		fmt.Println("#########################################################################")
		fmt.Printf("\n")
		c.cmdDir = dir
		err := c.Init(false)
		if err != nil {
			return err
		}

		ws, err := c.workspaces()
		if err != nil {
			return err
		}

		if contains(ws, "live-1") {
			log.WithFields(log.Fields{"dir": dir}).Info("Using live-1 context with: KUBE_CTX=live-1.cloud-platform.service.justice.gov.uk")

			c.cmdEnv = append(os.Environ(), "KUBE_CTX=live-1.cloud-platform.service.justice.gov.uk")
			c.Workspace = "live-1"

			err := c.Apply()
			if err != nil {
				return err
			}
		} else if contains(ws, "manager") {
			log.WithFields(log.Fields{"dir": dir}).Info("Using manager context with: KUBE_CTX=manager.cloud-platform.service.justice.gov.uk")

			c.cmdEnv = append(os.Environ(), "KUBE_CTX=manager.cloud-platform.service.justice.gov.uk")
			c.Workspace = "manager"

			err := c.Apply()
			if err != nil {
				return err
			}
		} else {
			log.WithFields(log.Fields{"dir": dir}).Info("No context, normal terraform plan")
			err := c.Apply()
			if err != nil {
				return err
			}
		}
	}

	return nil
}
