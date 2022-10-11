// Taken from: https://github.com/google/ko/blob/master/pkg/commands/version.go

package commands

import (
	"fmt"
	"runtime/debug"

	release "github.com/ministryofjustice/cloud-platform-cli/pkg/githubRelease/release"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// This value is set at build time. If it doesn't equal the latest
// version, the user will be prompted to upgrade.
// To build your binary you must pass
// the -ldflags "-X github.com/ministryofjustice/cloud-platform-cli/pkg/commands.Version=<version>"
var (
	Version      = "testBuild"
	Commit, Date string
)

const (
	owner      = "ministryofjustice"
	repoName   = "cloud-platform-cli"
	binaryName = "cloud-platform"
)

func addVersion(topLevel *cobra.Command) {
	topLevel.AddCommand(&cobra.Command{
		Use:    "version",
		Hidden: true,
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
			return "testBuild"
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
