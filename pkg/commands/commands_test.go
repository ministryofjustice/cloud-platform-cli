package commands_test

import (
	"flag"
	"os"
	"os/exec"
	"testing"

	"github.com/google/go-cmdtest"
)

var update = flag.Bool("update", false, "update test files with results")

func TestCommandOutput(t *testing.T) {
	ts, err := cmdtest.Read("testData")
	if err != nil {
		panic(err)
	}
	if err := exec.Command("go", "build", "-o", "cloud-platform", "../../.").Run(); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("cloud-platform")

	ts.Commands["cloud-platform"] = cmdtest.Program("cloud-platform")
	ts.Run(t, *update)
}
