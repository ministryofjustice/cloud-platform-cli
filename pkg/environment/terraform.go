package environment

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/kelseyhightower/envconfig"
)

const TerraformVersion = "0.14.6"

type Terraformer interface {
	Initialize()
	TerraformInitAndApply(namespace string, directory string) (map[string]string, error)
	TerraformDestroy(directory string) error
}

type TerraformerImpl struct {
	terraformBinaryPath string
	terraformVersion    string
	config              EnvBackendConfigVars
}

type EnvBackendConfigVars struct {
	PipelineStateBucket             string `required:"true" split_words:"true"`
	PipelineStateKeyPrefix          string `required:"true" split_words:"true"`
	PipelineTerraformStateLockTable string `required:"true" split_words:"true"`
	PipelineStateRegion             string `required:"true" split_words:"true"`
	PipelineCluster                 string `required:"true" split_words:"true"`
	PipelineClusterState            string `required:"true" split_words:"true"`
}

func NewTerraformer(terraformBinaryPath string) Terraformer {

	tf := TerraformerImpl{
		terraformVersion:    TerraformVersion,
		terraformBinaryPath: terraformBinaryPath,
	}
	tf.Initialize()
	return &tf
}

func (m *TerraformerImpl) Initialize() {
	var tfConfig EnvBackendConfigVars
	err := envconfig.Process("", &tfConfig)
	if err != nil {
		log.Fatalln("Terraform environment variables not set:", err.Error())
	}
	m.optionEnvBackendConfigVars(tfConfig)
}

func (m *TerraformerImpl) optionEnvBackendConfigVars(c EnvBackendConfigVars) error {
	m.config.PipelineStateBucket = c.PipelineStateBucket
	m.config.PipelineStateKeyPrefix = c.PipelineStateKeyPrefix
	m.config.PipelineTerraformStateLockTable = c.PipelineTerraformStateLockTable
	m.config.PipelineStateRegion = c.PipelineStateRegion
	m.config.PipelineCluster = c.PipelineCluster
	m.config.PipelineClusterState = c.PipelineClusterState
	return nil
}

func (m *TerraformerImpl) TerraformInitAndApply(namespace, directory string) (map[string]string, error) {
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

func (m *TerraformerImpl) TerraformDestroy(directory string) error {
	terraform, err := tfexec.NewTerraform(directory, m.terraformBinaryPath)
	if err != nil {
		return err
	}

	return terraform.Destroy(context.Background())
}
