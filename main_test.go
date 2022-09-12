package main_test

import (
	"flag"
	"os"
	"os/exec"
	"testing"

	"github.com/google/go-cmdtest"
)

var update = flag.Bool("update", false, "update test files with results")

func TestCmdOutput(t *testing.T) {
	if err := exec.Command("go", "build", "-o", "cloud-platform", ".").Run(); err != nil {
		t.Fatalf("Unable to build the binary %s", err)
	}
	defer os.Remove("cloud-platform")

	ts, err := cmdtest.Read("testData")
	if err != nil {
		t.Fatal(err)
	}
	ts.Commands["cloud-platform"] = cmdtest.Program("cloud-platform")
	ts.Run(t, *update)
}
