package deleteutils

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/terraform"
)

type TFDataAccessLayer interface {
	Init(context.Context, io.Writer) error
	Plan(context.Context, io.Writer) (bool, error)
	Destroy(context.Context, io.Writer) error
	WorkspaceDelete(context.Context, string) error
}

func InitTfCLI(tf *terraform.TerraformCLIConfig, dryRun bool) (TFDataAccessLayer, error) {
	if dryRun {
		tf.PlanVars = append(tf.PlanVars, tfexec.Destroy(true))
	}

	terraform, err := terraform.NewTerraformCLI(tf)
	if err != nil {
		return nil, err
	}

	return terraform, nil
}

func terraformInit(tf TFDataAccessLayer, workingDir string) error {
	// Start fresh and remove any local state.
	if err := terraform.DeleteLocalState(workingDir, ".terraform", ".terraform.lock.hcl"); err != nil {
		fmt.Println("Failed to delete local state, continuing anyway")
	}

	err := tf.Init(context.TODO(), log.Writer())
	if err != nil {
		return fmt.Errorf("failed to init terraform: %w", err)
	}

	return nil
}

func terraformDestroy(terraform TFDataAccessLayer, dryRun bool) error {
	if dryRun {
		if _, err := terraform.Plan(context.TODO(), log.Writer()); err != nil {
			return fmt.Errorf("destroy plan terraform failed: %w", err)
		}
	} else {
		if err := terraform.Destroy(context.TODO(), log.Writer()); err != nil {
			return fmt.Errorf("failed to destroy terraform: %w", err)
		}
	}
	return nil
}

func TerraformDestroyLayer(tf *terraform.TerraformCLIConfig, dryRun bool) error {
	tfCli, err := InitTfCLI(tf, dryRun)
	if err != nil {
		return err
	}

	err = terraformInit(tfCli, tf.WorkingDir)
	if err != nil {
		return err
	}

	err = terraformDestroy(tfCli, dryRun)
	if err != nil {
		return err
	}

	return nil
}
