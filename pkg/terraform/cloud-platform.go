package terraform

import (
	"errors"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

// TerraformWrapper calls Plan() and Apply() functions but ensures and check
// it is running using the right kubectl context
func (c *Commander) terraformWrapper(ws, dir, action string) error {
	// terraform init
	err := c.Init(false)
	if err != nil {
		return err
	}

	// terraform workspace select $ws
	c.Workspace = ws
	err = c.SelectWs(ws)
	if err != nil {
		return err
	}

	// Setting the right kubectl context
	log.WithFields(log.Fields{"dir": dir}).Info(fmt.Sprintf("Using %s context with: KUBE_CTX=%s.cloud-platform.service.justice.gov.uk", ws, ws))
	log.WithFields(log.Fields{"dir": dir}).Info(fmt.Sprintf("Setting kubectl current-context with: kubectl config use-context %s.cloud-platform.service.justice.gov.uk", ws))

	c.cmdEnv = append(os.Environ(), fmt.Sprintf("KUBE_CTX=%s.cloud-platform.service.justice.gov.uk", ws))
	err = c.kubectlUseContext(fmt.Sprintf("%s.cloud-platform.service.justice.gov.uk", ws))
	if err != nil {
		return err
	}

	cluster, err := c.kubectlContext()
	if err != nil {
		return err
	}

	// If we are not in the right kubectl context, we should be abort
	if cluster != ws {
		log.WithFields(log.Fields{"dir": dir}).Fatal("No matching Workspace and kubectl context")
		return errors.New("No matching Workspace and kubectl context in")
	}

	log.WithFields(log.Fields{"cluster": cluster, "workspace": c.Workspace, "dir": dir}).Info(fmt.Sprintf("Executing terraform %s", action))

	if action == "plan" {
		err = c.Plan()
		if err != nil {
			return err
		}
	}

	if action == "apply" {
		err = c.Apply()
		if err != nil {
			return err
		}
	}

	return nil
}

// BulkPlan executes plan against all directories that changed in the PR.
func (c *Commander) BulkPlan() error {
	dirs, err := targetDirs(c.BulkTfPaths)
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		prettyPrint(fmt.Sprintf("PLAN FOR DIRECTORY: %s\n", dir))
		c.cmdDir = dir
		err := c.Init(false)
		if err != nil {
			return err
		}

		ws, err := c.workspaces()
		if err != nil {
			return err
		}

		if contains(ws, "live") {
			c.terraformWrapper("live", dir, "plan")
		}

		if contains(ws, "manager") {
			c.terraformWrapper("manager", dir, "plan")
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
		prettyPrint(fmt.Sprintf("APPLY FOR DIRECTORY: %s\n", dir))
		c.cmdDir = dir
		err := c.Init(false)
		if err != nil {
			return err
		}

		ws, err := c.workspaces()
		if err != nil {
			return err
		}

		if contains(ws, "live") {
			c.terraformWrapper("live", dir, "apply")
		}

		if contains(ws, "manager") {
			c.terraformWrapper("manager", dir, "apply")
		}

	}

	return nil
}
