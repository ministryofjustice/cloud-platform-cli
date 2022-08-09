package environment

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os/exec"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/kelseyhightower/envconfig"
)

const TerraformVersion = "0.14.6"

type Applier interface {
	Initialize()
	KubectlApply(directory string) (string, error)
	TerraformInitAndApply(namespace string, directory string) (map[string]string, error)
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
		log.Fatalln("Terraform and Kubeconfig environment variables not set:", err.Error())
	}
	m.optionEnvBackendConfigVars(tfConfig)
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

func (m *ApplierImpl) TerraformInitAndApply(namespace, directory string) (map[string]string, error) {
	terraform, err := tfexec.NewTerraform(directory, m.terraformBinaryPath)
	if err != nil {
		return map[string]string{}, errors.New("unable to instantiate Terraform: " + err.Error())
	}

	key := m.config.PipelineStateKeyPrefix + m.config.PipelineClusterState + "/" + namespace + "/terraform.tfstate"

	err = terraform.Init(context.Background(),
		tfexec.BackendConfig(fmt.Sprintf("bucket=%s", m.config.PipelineStateBucket)),
		tfexec.BackendConfig(fmt.Sprintf("key=%s", key)),
		tfexec.BackendConfig(fmt.Sprintf("dynamodb_table=%s", m.config.PipelineTerraformStateLockTable)),
		tfexec.BackendConfig(fmt.Sprintf("region=%s", m.config.PipelineStateRegion)))
	if err != nil {
		return nil, err
	}

	log.Println("Applying Terraform")
	err = terraform.Apply(context.Background(), tfexec.Refresh(false))
	if err != nil {
		return nil, errors.New("unable to apply Terraform: " + err.Error())
	}

	rawOutputs, _ := terraform.Output(context.Background())
	outputs := make(map[string]string, len(rawOutputs))
	for outputName, outputRawValue := range rawOutputs {
		outputValue := string(outputRawValue.Value)
		// Strip the first and last quote which gets added for some reason
		outputValue = outputValue[1 : len(outputValue)-1]
		outputs[outputName] = outputValue
	}
	return outputs, nil
}

func (m *ApplierImpl) TerraformDestroy(directory string) error {
	terraform, err := tfexec.NewTerraform(directory, m.terraformBinaryPath)
	if err != nil {
		return err
	}

	return terraform.Destroy(context.Background())
}

func (m *ApplierImpl) KubectlApply(directory string) (string, error) {
	args := []string{"kubectl", "apply", "-f", directory}

	stdout, err := exec.Command(args[0], args[1:]...).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("Error: %v", err)
	}

	return string(stdout), err
}
