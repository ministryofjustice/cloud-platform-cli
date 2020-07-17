package environment

import (
	"testing"
)

func TestRunningOutsideEnvironmentsWorkingCopy(t *testing.T) {
	err := CreateTemplateNamespace(nil, nil)
	if err.Error() != "You are outside cloud-platform-environment repo" {
		t.Errorf("Unexpected error: %s", err)
	}
}
