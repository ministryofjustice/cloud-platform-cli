package terraform

import (
	"bytes"
	"os/exec"
	"syscall"

	log "github.com/sirupsen/logrus"
)

// Commander empty struct which methods to execute terraform
type Commander struct {
	action string
	cmd    []string
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
	})

	cmd := exec.Command("terraform", args...)

	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
		}
		contextLogger.Error("cmd.Run() failed")
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
func (s *Commander) CheckDivergence(ws string, cmd ...string) error {

	err := s.Init()
	if err != nil {
		return err
	}

	err = s.SelectWs(ws)
	if err != nil {
		return err
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
	log.Info(output.Stdout)

	if output.ExitCode == 0 {
		return nil
	}

	return err
}
