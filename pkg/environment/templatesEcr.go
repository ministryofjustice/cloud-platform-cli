package environment

import (
	"fmt"

	"github.com/spf13/cobra"
)

func CreateTemplateEcr(cmd *cobra.Command, args []string) error {
	fmt.Println("I will create an ECR for you.")
	return nil
}
