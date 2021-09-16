package environment

import "testing"

func TestMigrateCheck(t *testing.T) {
	randomNS := "mogaal-namespace-that-doesnt-exist"

	err := MigrateCheck(randomNS)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
}
