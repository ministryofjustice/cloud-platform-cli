package environment

import (
	"fmt"

	otiai10 "github.com/otiai10/copy"
	"github.com/spf13/cobra"
)

func Migrate(cmd *cobra.Command, args []string) error {
	re := RepoEnvironment{}

	// this already checks we are within the environment repo.
	err := re.mustBeInANamespaceFolder()
	if err != nil {
		return err
	}

	nsName, err := re.getNamespaceName()
	if err != nil {
		return err
	}

	src := fmt.Sprintf("../%s", nsName)
	dst := fmt.Sprintf("../../live.cloud-platform.service.justice.gov.uk/%s", nsName)

	err = otiai10.Copy(src, dst)
	if err != nil {
		return err
	}

	fmt.Printf("Namespace %s was succesffully migrated to live folder", nsName)

	return nil
}
