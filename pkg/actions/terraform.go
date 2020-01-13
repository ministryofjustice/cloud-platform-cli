package actions

import (
	"fmt"

	terraform "github.com/ministryofjustice/cloud-platform-tools/pkg/terraform"
	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

// TerraformCheckDivergence tell us if there is a divergence in our terraform plan.
// This basically translates in: "is there any changes pending to apply?"
func TerraformCheckDivergence(c *cli.Context) error {

	contextLogger := log.WithFields(log.Fields{"subcommand": "check-divergence"})

	workspace := c.String("workspace")
	terraform := terraform.Commander{}
	var TfVarFile []string

	// Check if user provided a terraform var-file
	if c.String("var-file") != "" {
		TfVarFile = append([]string{fmt.Sprintf("-var-file=%s", c.String("var-file"))})
	}

	contextLogger.Info("Executing terraform plan, if there is a drift this program execution will fail")

	err := terraform.CheckDivergence(workspace, TfVarFile...)
	if err != nil {
		contextLogger.Error("Error executing plan, either an error or a divergence")
		return err
	}

	return nil
}
