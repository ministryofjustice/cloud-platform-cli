package terraform

import (
	"bytes"
	"os/exec"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

// Kubectl creates kubectl command to be executed
func (s *Commander) Kubectl(args ...string) (*CmdOutput, error) {

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

	cmd := exec.Command("kubectl", args...)

	if s.cmdDir != "" {
		cmd.Dir = s.cmdDir
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

// kubectlContext return the cluster name of current context set in kubectl.
// PS: It is assumed the context is stored in format <cluster-name>.cloud-platform.service.justice.gov.uk
// in kubeconfig file from where kubectl fetches the context.
func (c *Commander) kubectlContext() (string, error) {
	arg := []string{
		"config",
		"current-context",
	}

	output, err := c.Kubectl(arg...)
	if err != nil {
		return "", err
	}

	cluster := strings.Split(output.Stdout, ".")
	return cluster[0], nil
}

// kubectlUseContext use the context passed as input arg.
func (c *Commander) kubectlUseContext(contextName string) error {
	arg := []string{
		"config",
		"use-context",
		contextName,
	}

	_, err := c.Kubectl(arg...)
	if err != nil {
		return err
	}

	return nil
}
