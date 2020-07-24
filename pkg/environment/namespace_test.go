package environment

import (
	"testing"
)

// If we assign a string value to 'Namespace', we get it back
func TestNamespaceName(t *testing.T) {
	ns := Namespace{Namespace: "foobar"}
	if ns.Namespace != "foobar" {
		t.Errorf("Something went wrong: %s", ns.Namespace)
	}
}

// Test getting namespace information from a file
func TestNamespaceFromYamlFile(t *testing.T) {
	ns := Namespace{}
	ns.readYamlFile("fixtures/foobar-namespace.yml")
	if ns.Namespace != "foobar" {
		t.Errorf("Expect foobar, got: %s", ns.Namespace)
	}

	if ns.IsProduction != "false" {
		t.Errorf("Expect foobar, got: %s", ns.IsProduction)
	}

	if ns.BusinessUnit != "MoJ Digital" {
		t.Errorf("Expect foobar, got: %s", ns.BusinessUnit)
	}

	if ns.Owner != "Cloud Platform: david.salgado@digital.justice.gov.uk" {
		t.Errorf("Expect foobar, got: %s", ns.Owner)
	}

	if ns.OwnerEmail != "david.salgado@digital.justice.gov.uk" {
		t.Errorf("Expect foobar, got: %s", ns.OwnerEmail)
	}

	if ns.Environment != "development" {
		t.Errorf("Expect foobar, got: %s", ns.Environment)
	}

	if ns.Application != "David Salgado test namespace" {
		t.Errorf("Expect foobar, got: %s", ns.Application)
	}

	if ns.SourceCode != "https://github.com/ministryofjustice/cloud-platform-environments" {
		t.Errorf("Expect foobar, got: %s", ns.SourceCode)
	}
}
