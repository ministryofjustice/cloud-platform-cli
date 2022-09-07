package test_test

import (
	"fmt"
	"testing"

	cloudplatform "github.com/ministryofjustice/cloud-platform-cli/cmd"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var t *testing.T
var root = &cobra.Command{Use: "root", RunE: cloudplatform.RootCmdRunE}
var _ = Describe("Version", func() {
	It("should be 1.0.0", func() {
		// TODO: Test test - please remove
		Expect("1.0.0").To(Equal("1.0.0"))
	})
	It("Should return the testBuild string", func() {
		fmt.Println("Start testBuild")
		// version, err := cloudplatform.ExecuteCommand(t, root, "")
		cloudplatform.RootCmdFlags(root)
		cloudplatform.Execute()
		// if err != nil {
		// 	log.Fatalln(err)
		// }
		// fmt.Println(version)
	})
})
