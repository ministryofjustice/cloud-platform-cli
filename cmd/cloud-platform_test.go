package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func ExecuteCommand(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)

	args = append([]string{"--skip-version-check"}, args...)

	err := cmd.Execute()
	return strings.TrimSpace(buf.String()), err
}

func TestCloudPlatformRootCmd(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		wantOut string
	}{
		{
			name:    "running the cloud-platform command with no args",
			args:    nil,
			wantErr: false,
			wantOut: rootCmd.Short,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := ExecuteCommand(t, rootCmd, tc.args...)
			if (err != nil) != tc.wantErr {
				t.Errorf("execute() \nerror = \n%v \nwantErr \n%v", err, tc.wantErr)
				return
			}
			// Trim whitespace from the output
			out = strings.TrimSpace(out)
			if !strings.Contains(out, tc.wantOut) {
				t.Errorf("execute() \ngot = \n%v \nwant = \n%v", out, tc.wantOut)
			}
		})
	}
}

// package cmd

// import (
//   "errors"

//   "github.com/spf13/cobra"
// )

// var rootCmd = &cobra.Command{
//   Use:  "example",
//   RunE: RootCmdRunE,
// }

// func RootCmdRunE(cmd *cobra.Command, args []string) error {
//   t, err := cmd.Flags().GetBool("toggle")

//   if err != nil {
//     return err
//   }

//   if t {
//     cmd.Println("ok")
//     return nil
//   }

//   return errors.New("not ok")
// }

// func RootCmdFlags(cmd *cobra.Command) {
//   cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
// }

// func Execute() {
//   cobra.CheckErr(rootCmd.Execute())
// }

// func init() {
//   RootCmdFlags(rootCmd)
// }
