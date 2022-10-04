package environment

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/kelseyhightower/envconfig"
)

const TerraformVersion = "0.14.8"

type Applier interface {
	Initialize()
	KubectlApply(namespace, directory string, dryRun bool) (string, error)
	ConfigureInit(namespace string) []tfexec.InitOption
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

func (m *ApplierImpl) ConfigureInit(namespace string) []tfexec.InitOption {

	key := m.config.PipelineStateKeyPrefix + m.config.PipelineClusterState + "/" + namespace + "/terraform.tfstate"

	InitVars := make([]tfexec.InitOption, 0)
	InitVars = append(InitVars, tfexec.BackendConfig(fmt.Sprintf("bucket=%s", m.config.PipelineStateBucket)))
	InitVars = append(InitVars, tfexec.BackendConfig(fmt.Sprintf("key=%s", key)))
	InitVars = append(InitVars, tfexec.BackendConfig(fmt.Sprintf("dynamodb_table=%s", m.config.PipelineTerraformStateLockTable)))
	InitVars = append(InitVars, tfexec.BackendConfig(fmt.Sprintf("region=%s", m.config.PipelineStateRegion)))

	return InitVars
}

func (m *ApplierImpl) KubectlApply(namespace, directory string, dryRun bool) (string, error) {
	var args []string
	if dryRun {
		args = []string{"kubectl", "-n", namespace, "apply", "--dry-run", "-f", directory}
	} else {
		args = []string{"kubectl", "-n", namespace, "apply", "-f", directory}
	}

	stdout, err := exec.Command(args[0], args[1:]...).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("error: %v", err)
	}

	return string(stdout), err
}
