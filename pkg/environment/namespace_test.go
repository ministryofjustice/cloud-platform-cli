package environment

import (
	"testing"
)

// If we assign a string value to 'name', we get it back
func TestNamespaceName(t *testing.T) {
	ns := Namespace{name: "foobar"}
	if ns.name != "foobar" {
		t.Errorf("Something went wrong: %s", ns.name)
	}
}

// Test getting namespace information from a file
func TestNamespaceFromYamlFile(t *testing.T) {
	ns := Namespace{}
	ns.readYamlFile("fixtures/foobar-namespace.yml")
	if ns.name != "foobar" {
		t.Errorf("Expect foobar, got: %s", ns.name)
	}

	if ns.isProduction != "false" {
		t.Errorf("Expect foobar, got: %s", ns.isProduction)
	}

	if ns.businessUnit != "MoJ Digital" {
		t.Errorf("Expect foobar, got: %s", ns.businessUnit)
	}

	if ns.owner != "Cloud Platform: david.salgado@digital.justice.gov.uk" {
		t.Errorf("Expect foobar, got: %s", ns.owner)
	}

	if ns.ownerEmail != "david.salgado@digital.justice.gov.uk" {
		t.Errorf("Expect foobar, got: %s", ns.ownerEmail)
	}

	if ns.environmentName != "development" {
		t.Errorf("Expect foobar, got: %s", ns.environmentName)
	}

	if ns.application != "David Salgado test namespace" {
		t.Errorf("Expect foobar, got: %s", ns.application)
	}

	if ns.sourceCode != "https://github.com/ministryofjustice/cloud-platform-environments" {
		t.Errorf("Expect foobar, got: %s", ns.sourceCode)
	}
}
