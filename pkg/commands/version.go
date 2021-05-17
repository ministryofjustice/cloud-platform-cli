// Taken from: https://github.com/google/ko/blob/master/pkg/commands/version.go

package commands

import (
	"fmt"
	"runtime/debug"

	release "github.com/ministryofjustice/cloud-platform-cli/pkg/github/release"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// This MUST match the number of the latest release on github
var Version = "1.9.5"

const owner = "ministryofjustice"
const repoName = "cloud-platform-cli"
const binaryName = "cloud-platform"

func addVersion(topLevel *cobra.Command) {
	topLevel.AddCommand(&cobra.Command{
		Use:    "version",
		Short:  `Print version`,
		PreRun: upgradeIfNotLatest,
		Run: func(cmd *cobra.Command, args []string) {
			v := version()
			if v == "" {
				fmt.Println("could not determine build information")
			} else {
				fmt.Println(v)
			}
		},
	})
}

func version() string {
	if Version == "" {
		i, ok := debug.ReadBuildInfo()
		if !ok {
			return ""
		}
		Version = i.Main.Version
	}
	return Version
}

func upgradeIfNotLatest(cmd *cobra.Command, args []string) {
	if viper.GetBool("skip-version-check") {
		return
	}
	r := release.New(owner, repoName, Version, binaryName)
	r.UpgradeIfNotLatest()
}
