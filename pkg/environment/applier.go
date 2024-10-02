package environment

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os/exec"

	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/kelseyhightower/envconfig"
)

const TerraformVersion = "1.2.5"

type Applier interface {
	Initialize()
	KubectlApply(namespace, directory string, dryRun bool) (string, error)
	KubectlDelete(namespace, directory string, dryRun bool) (string, error)
	TerraformInitAndPlan(namespace string, directory string) (*tfjson.Plan, string, error)
	TerraformInitAndApply(namespace string, directory string) (string, error)
	TerraformInitAndDestroy(namespace string, directory string) (string, error)
	TerraformDestroy(directory string) error
}

type ApplierImpl struct {
	terraformBinaryPath string
	kubectlBinaryPath   string
	terraformVersion    string
	config              EnvBackendConfigVars
}

type EnvBackendConfigVars struct {
	kubeconfig                      string `required:"true"`
	PipelineStateBucket             string `required:"true" split_words:"true"`
	PipelineStateKeyPrefix          string `required:"true" split_words:"true"`
	PipelineTerraformStateLockTable string `required:"true" split_words:"true"`
	PipelineStateRegion             string `required:"true" split_words:"true"`
	PipelineCluster                 string `required:"true" split_words:"true"`
	PipelineClusterState            string `required:"true" split_words:"true"`
}

func NewApplier(terraformBinaryPath string, kubectlBinaryPath string) Applier {
	applier := ApplierImpl{
		terraformVersion:    TerraformVersion,
		terraformBinaryPath: terraformBinaryPath,
		kubectlBinaryPath:   kubectlBinaryPath,
	}
	applier.Initialize()
	return &applier
}

func (m *ApplierImpl) Initialize() {
	var tfConfig EnvBackendConfigVars
	err := envconfig.Process("", &tfConfig)
	if err != nil {
		log.Fatalln("Terraform backend and Kubeconfig environment variables not set:", err.Error())
	}
	err = m.optionEnvBackendConfigVars(tfConfig)
	if err != nil {
		log.Fatalln("Terraform backend and Kubeconfig environment variables not set:", err.Error())
	}
}

func (m *ApplierImpl) optionEnvBackendConfigVars(c EnvBackendConfigVars) error {
	m.config.PipelineStateBucket = c.PipelineStateBucket
	m.config.PipelineStateKeyPrefix = c.PipelineStateKeyPrefix
	m.config.PipelineTerraformStateLockTable = c.PipelineTerraformStateLockTable
	m.config.PipelineStateRegion = c.PipelineStateRegion
	m.config.PipelineCluster = c.PipelineCluster
	m.config.PipelineClusterState = c.PipelineClusterState
	m.config.kubeconfig = c.kubeconfig
	return nil
}

func (m *ApplierImpl) TerraformInitAndApply(namespace, directory string) (string, error) {
	var out bytes.Buffer

	terraform, err := tfexec.NewTerraform(directory, m.terraformBinaryPath)
	if err != nil {
		return "", errors.New("unable to instantiate Terraform: " + err.Error())
	}

	terraform.SetStdout(&out)
	terraform.SetStderr(&out)

	// Sometimes the error text would be useful in the command output that's
	// displayed in the UI. For this reason, we append the error to the
	// output before we return it.
	errReturn := func(out bytes.Buffer, err error) (string, error) {
		if err != nil {
			return fmt.Sprintf("%s\n%s", out.String(), err.Error()), err
		}

		return out.String(), nil
	}

	key := m.config.PipelineStateKeyPrefix + m.config.PipelineClusterState + "/" + namespace + "/terraform.tfstate"

	err = terraform.Init(context.Background(),
		tfexec.BackendConfig(fmt.Sprintf("bucket=%s", m.config.PipelineStateBucket)),
		tfexec.BackendConfig(fmt.Sprintf("key=%s", key)),
		tfexec.BackendConfig(fmt.Sprintf("dynamodb_table=%s", m.config.PipelineTerraformStateLockTable)),
		tfexec.BackendConfig(fmt.Sprintf("region=%s", m.config.PipelineStateRegion)))
	if err != nil {
		return errReturn(out, err)
	}

	err = terraform.Apply(context.Background(), tfexec.Refresh(true))
	if err != nil {
		return errReturn(out, err)
	}

	return out.String(), nil
}

func (m *ApplierImpl) TerraformInitAndPlan(namespace, directory string) (*tfjson.Plan, string, error) {
	var out bytes.Buffer
	terraform, err := tfexec.NewTerraform(directory, m.terraformBinaryPath)
	if err != nil {
		return nil, "", errors.New("unable to instantiate Terraform: " + err.Error())
	}

	terraform.SetStdout(&out)
	terraform.SetStderr(&out)

	// Sometimes the error text would be useful in the command output that's
	// displayed in the UI. For this reason, we append the error to the
	// output before we return it.
	errReturn := func(out bytes.Buffer, err error) (*tfjson.Plan, string, error) {
		if err != nil {
			return nil, fmt.Sprintf("%s\n%s", out.String(), err.Error()), err
		}

		return nil, out.String(), nil
	}

	key := m.config.PipelineStateKeyPrefix + m.config.PipelineClusterState + "/" + namespace + "/terraform.tfstate"

	err = terraform.Init(context.Background(),
		tfexec.BackendConfig(fmt.Sprintf("bucket=%s", m.config.PipelineStateBucket)),
		tfexec.BackendConfig(fmt.Sprintf("key=%s", key)),
		tfexec.BackendConfig(fmt.Sprintf("dynamodb_table=%s", m.config.PipelineTerraformStateLockTable)),
		tfexec.BackendConfig(fmt.Sprintf("region=%s", m.config.PipelineStateRegion)))
	if err != nil {
		return errReturn(out, err)
	}

	outOption := tfexec.Out("plan-" + namespace + ".out")
	_, err = terraform.Plan(context.Background(), outOption)

	tfPlan, _ := terraform.ShowPlanFile(context.Background(), "plan-"+namespace+".out")

	tfPlan.UseJSONNumber(true)

	if err != nil {
		return nil, "", errors.New("unable to do Terraform Plan: " + err.Error())
	}

	// ignore if any changes or no changes.
	_, err = terraform.Plan(context.Background())
	if err != nil {
		return nil, "", errors.New("unable to do Terraform Plan: " + err.Error())
	}

	return tfPlan, out.String(), nil
}

func (m *ApplierImpl) TerraformInitAndDestroy(namespace, directory string) (string, error) {
	var out bytes.Buffer
	terraform, err := tfexec.NewTerraform(directory, m.terraformBinaryPath)
	if err != nil {
		return "", errors.New("unable to instantiate Terraform: " + err.Error())
	}

	terraform.SetStdout(&out)
	terraform.SetStderr(&out)

	// Sometimes the error text would be useful in the command output that's
	// displayed in the UI. For this reason, we append the error to the
	// output before we return it.
	errReturn := func(out bytes.Buffer, err error) (string, error) {
		if err != nil {
			return fmt.Sprintf("%s\n%s", out.String(), err.Error()), err
		}

		return out.String(), nil
	}

	key := m.config.PipelineStateKeyPrefix + m.config.PipelineClusterState + "/" + namespace + "/terraform.tfstate"

	err = terraform.Init(context.Background(),
		tfexec.BackendConfig(fmt.Sprintf("bucket=%s", m.config.PipelineStateBucket)),
		tfexec.BackendConfig(fmt.Sprintf("key=%s", key)),
		tfexec.BackendConfig(fmt.Sprintf("dynamodb_table=%s", m.config.PipelineTerraformStateLockTable)),
		tfexec.BackendConfig(fmt.Sprintf("region=%s", m.config.PipelineStateRegion)))
	if err != nil {
		return errReturn(out, err)
	}

	// ignore if any changes or no changes.
	err = terraform.Destroy(context.Background())
	if err != nil {
		return "", errors.New("unable to do Terraform Destroy: " + err.Error())
	}

	return out.String(), nil
}

func (m *ApplierImpl) TerraformDestroy(directory string) error {
	terraform, err := tfexec.NewTerraform(directory, m.terraformBinaryPath)
	if err != nil {
		return err
	}

	return terraform.Destroy(context.Background())
}

func (m *ApplierImpl) KubectlApply(namespace, directory string, dryRun bool) (string, error) {
	var args []string
	if dryRun {
		args = []string{"kubectl", "-n", namespace, "apply", "--dry-run=client", "-f", directory}
	} else {
		args = []string{"kubectl", "-n", namespace, "apply", "-f", directory}
	}

	stdout, err := exec.Command(args[0], args[1:]...).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("error: %v", err)
	}

	return string(stdout), err
}

func (m *ApplierImpl) KubectlDelete(namespace, directory string, dryRun bool) (string, error) {
	var args []string
	if dryRun {
		args = []string{"kubectl", "-n", namespace, "delete", "--dry-run=client", "-f", directory}
	} else {
		args = []string{"kubectl", "-n", namespace, "delete", "-f", directory}
	}

	stdout, err := exec.Command(args[0], args[1:]...).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("error: %v", err)
	}

	return string(stdout), err
}
