package environment

import (
	"os"
	"testing"
)

// TestCreateServiceAccountFiles simulates a ServiceAccount object
// and tests whether the createSvcAccFile function creates a serviceaccount
// manifest. To do this, the function requires there to be a namespace manifest
// called 00-namespace.yaml
func TestCreateServiceAccountFiles(t *testing.T) {
	svc := ServiceAccount{
		Name:      "testName",
		Namespace: "foobar",
	}

	svcAccFileName := "05-serviceaccount.yaml"
	svcAccNamespace := "00-namespace.yaml"

	err := os.Link("fixtures/foobar-namespace.yml", svcAccNamespace)
	if err != nil {
		t.Error("Error copying file:", err)
	}
	err = svc.createSvcAccFile(svc.Name)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	fileContainsString(t, svcAccFileName, svc.Name)
	fileContainsString(t, svcAccFileName, svc.Namespace)

	os.Remove(svcAccFileName)
	os.Remove(svcAccNamespace)
}
