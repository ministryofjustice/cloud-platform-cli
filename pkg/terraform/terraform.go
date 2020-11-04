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

// Commander empty struct which methods to execute terraform
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
	BulkTfPlanPaths string
}

// CmdOutput has the Stout and Stderr
type CmdOutput struct {
	Stdout   string
	Stderr   string
	ExitCode int
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
		contextLogger.Error("cmd.Run() failed")
		return nil, err
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

// Init is mandatory almost always before doing anything with terraform
func (s *Commander) Init() error {

	output, err := s.Terraform("init")
	if err != nil {
		log.Fatal("Error running terraform init")
	}

	log.Info(output.Stdout)

	return nil
}

// SelectWs is used to select certain workspace
func (s *Commander) SelectWs(ws string) error {

	output, err := s.Terraform("workspace", "select", ws)
	if err != nil {
		return err
	}

	log.Info(output.Stdout)

	return nil
}

// CheckDivergence is used to select certain workspace
func (s *Commander) CheckDivergence() error {
	err := s.Init()
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
			"-detailed-exitcode",
		},
		cmd...,
	)

	output, err := s.Terraform(arg...)

	if err != nil {
		log.Error(err)
		log.Error(output.Stderr)
	}

	if s.DisplayTfOutput {
		fmt.Println(output.Stdout)
	}

	if output.ExitCode == 0 {
		return nil
	}

	return err
}

// Apply executes terraform apply
func (s *Commander) Apply() error {
	err := s.Init()
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
		log.Error(err)
		log.Error(output.Stderr)
	}

	if s.DisplayTfOutput {
		fmt.Println(output.Stdout)
	}

	if output.ExitCode == 0 {
		return nil
	}

	return err

}

// Plan executes terraform apply
func (s *Commander) Plan() error {
	err := s.Init()
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
		log.Error(err)
		log.Error(output.Stderr)
	}

	if s.DisplayTfOutput {
		fmt.Println(output.Stdout)
	}

	if output.ExitCode == 0 {
		return nil
	}

	return err

}

func (c *Commander) workspaces() ([]string, error) {
	arg := []string{
		"workspace",
		"list",
	}

	output, err := c.Terraform(arg...)
	if err != nil {
		log.Error(output.Stderr)
		return nil, err
	}

	ws := strings.Split(output.Stdout, "\n")

	return ws, nil
}

func (c *Commander) BulkPlan() error {
	dirs, err := targetDirs(c.BulkTfPlanPaths)
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		c.cmdDir = dir
		err := c.Init()
		if err != nil {
			return err
		}

		ws, err := c.workspaces()
		if err != nil {
			return err
		}

		if contains(ws, "  live-1") {
			log.WithFields(log.Fields{"dir": dir}).Info("Using live-1 context with: KUBE_CTX=live-1.cloud-platform.service.justice.gov.uk")

			c.cmdEnv = append(os.Environ(), "KUBE_CTX=live-1.cloud-platform.service.justice.gov.uk")
			c.Workspace = "live-1"

			err := c.Plan()
			if err != nil {
				return err
			}
		} else if contains(ws, "  manager") {
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
