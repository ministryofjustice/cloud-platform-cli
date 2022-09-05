package cmd

import (
	"testing"

	"github.com/matryer/is"
)

func TestCloudPlatformRootCmd(t *testing.T) {
	is := is.New(t)

	err := rootCmd.Execute()

	is.NoErr(err)
}
