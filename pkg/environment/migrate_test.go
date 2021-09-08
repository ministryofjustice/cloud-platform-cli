package environment

import "testing"

func TestGrepFile(t *testing.T) {
	hasBusinessUnit := grepFile("fixtures/foobar-namespace.yml", []byte("cloud-platform.justice.gov.uk/business-unit"))
	if hasBusinessUnit == 0 {
		t.Errorf("Business Unit annotation exist inside fixures file, grepFile() returned %v - expected: 1", hasBusinessUnit)
	}

	hasRandomAnnotation := grepFile("fixtures/foobar-namespace.yml", []byte("whatever"))
	if hasRandomAnnotation != 0 {
		t.Errorf("whatever annotation DOES NOT exist inside fixures file, grepFile() returned %v - expected: 0", hasRandomAnnotation)
	}
}
