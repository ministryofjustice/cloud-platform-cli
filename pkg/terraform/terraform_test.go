package terraform

import "testing"

func TestNewOptions(t *testing.T) {
	options, err := NewOptions("1.1.1", "testWorkspace")
	if err == nil {
		t.Errorf("newTerraformOptions() error = %v", "expected error")
	}

	if options.Version != "1.1.1" {
		t.Errorf("newTerraformOptions() error = %v", "terraform version not set")
	}
}
