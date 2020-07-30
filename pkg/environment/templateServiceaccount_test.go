package environment

import (
	"os"
	"testing"
)

func TestCreateServiceAccountFiles(t *testing.T) {
	svc := ServiceAccount{
		Name:      "testName",
		Namespace: "testNamespace",
	}

	fileName := "05-serviceaccount.yaml"

	err := svc.createSvcFile()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	fileContainsString(t, fileName, svc.Name)
	fileContainsString(t, fileName, svc.Namespace)

	os.Remove(fileName)
}
