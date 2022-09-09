package commands_test

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/google/go-cmdtest"
)

var update = flag.Bool("update", false, "update test files with results")

func TestOutputOfCmd(t *testing.T) {
	if err := exec.Command("go", "build", "-o", "cloud-platform", "../../.").Run(); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("cloud-platform")

	// for subdirectory in directory do
	dirs, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}
	var ts *cmdtest.TestSuite
	for _, dir := range dirs {
		fmt.Println(dir.Name())
		switch dir.Name() {
		case "environmentCmds":
			ts, err = createTestEnvironment(t, dir.Name())
			if err != nil {
				t.Fatal(err)
			}
		default:
			testSuite(t, dir.Name())
		}

		// ts, err := cmdtest.Read("testdata/" + dir.Name())
		// if err != nil {
		// 	t.Fatal(err)
		// }
		// ts.Commands["cloud-platform"] = cmdtest.Program("cloud-platform")
		ts.Run(t, *update)
	}
}

func testSuite(t *testing.T, dir string) (*cmdtest.TestSuite, error) {
	t.Helper()
	ts, err := cmdtest.Read("testdata/" + dir)
	if err != nil {
		return nil, err
	}
	ts.Commands["cloud-platform"] = cmdtest.Program("cloud-platform")
	return ts, nil
}

func createTestEnvironment(t *testing.T, dir string) (*cmdtest.TestSuite, error) {
	t.Helper()
	fmt.Println("createTestEnvironment")
	env := "cloud-platform-environments"
	ts, err := testSuite(t, dir)
	if err != nil {
		return nil, err
	}
	fmt.Println(ts)

	// create temp env
	ts.Commands["mkdir"] = cmdtest.Program("testdata/" + dir + env + "/namespaces" + "live.cloud-platform.service.justice.gov.uk")
	ts.Commands["chdir"] = cmdtest.Program("testdata/" + dir + env)
	ts.Setup = func(rootDir string) error {
		fmt.Println("Setup")
		if err := os.MkdirAll("testdata/"+dir+"/cloud-platform-environments", 0755); err != nil {
			return err
		}
		if err := os.Chdir("testdata/" + dir + "/cloud-platform-environments"); err != nil {
			return err
		}
		return nil
	}

	ts.KeepRootDirs = true
	// if err := os.MkdirAll("testdata/"+dir+"/cloud-platform-environments", 0755); err != nil {
	// 	return nil, err
	// }
	// change directory to the above
	// if err := os.Chdir("testdata/" + dir + "/cloud-platform-environments"); err != nil {
	// 	return nil, err
	// }

	// defer os.Chdir("../../..")
	// defer os.RemoveAll("testdata/" + dir + "/cloud-platform-environments")
	// create a test createTestEnvironment

	return ts, nil
}

// func TestBasicOutput(t *testing.T) {
// 	ts, err := cmdtest.Read("basicTestData")
// 	if err != nil {
// 		panic(err)
// 	}
// 	if err := exec.Command("go", "build", "-o", "cloud-platform", "../../.").Run(); err != nil {
// 		t.Fatal(err)
// 	}
// 	defer os.Remove("cloud-platform")

// 	ts.Commands["cloud-platform"] = cmdtest.Program("cloud-platform")
// 	ts.Run(t, *update)
// }

// func TestEnvironmentOutput(t *testing.T) {
// 	ts, err := cmdtest.Read("environmentTestData")
// 	if err != nil {
// 		panic(err)
// 	}
// 	if err := exec.Command("go", "build", "-o", "cloud-platform", "../../.").Run(); err != nil {
// 		t.Fatal(err)
// 	}
// 	defer os.Remove("cloud-platform")

// 	ts.Commands["cloud-platform"] = cmdtest.Program("cloud-platform")
// 	ts.Run(t, *update)
// }
