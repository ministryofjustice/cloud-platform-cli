package environment_test

import (
	"log"
	"os"
	"testing"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/environment"
)

var file *os.File

func TestMain(m *testing.M) {
	var err error
	file, err = createMockKubeConfigFile("temp")
	if err != nil {
		log.Fatalln(err)
	}
	code := m.Run()

	defer file.Close()
	defer os.Remove(file.Name())
	os.Exit(code)
}

func createMockKubeConfigFile(path string) (*os.File, error) {
	data := []byte(`
apiVersion: v1
clusters:
- cluster:
    server: https://127.0.0.1:55171
  name: kind-kind
- cluster:
    server: https://127.0.0.1:55902
  name: kind-kind2
contexts:
- context:
    cluster: kind-kind
    user: kind-kind
  name: kind-kind
- context:
    cluster: kind-kind2
    user: kind-kind2
  name: kind-kind2
current-context: kind-kind2
kind: Config
preferences: {}
users:
- name: kind-kind
  user:
- name: kind-kind2
  user:
`)

	file, err := os.CreateTemp("", "temp")
	if err != nil {
		return nil, err
	}

	if _, err := file.Write(data); err != nil {
		return nil, err
	}

	return file, nil
}

func TestNewDivergence(t *testing.T) {
	divergence, err := environment.NewDivergence("kind", file.Name(), "ghp_fake", nil)
	if err != nil {
		t.Fatalf("error creating divergence object, when it should have created: %v", err)
	}

	if divergence == nil {
		t.Fatalf("divergence is nil, when it should have created")
	}

	if divergence.ClusterName != "kind" {
		t.Fatalf("divergence cluster name is %s, when it should be kind", divergence.ClusterName)
	}
}
