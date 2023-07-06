package environment

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"

	"github.com/MakeNowJust/heredoc"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

const (
	rdsTemplateFilePrefix = "https://raw.githubusercontent.com/ministryofjustice/cloud-platform-terraform-rds-instance/main/examples/"
	rdsTfFilePrefix       = "resources/"
)

// CreateTemplateRds creates the terraform files from environment's template folder
func CreateTemplateRds(cmd *cobra.Command, args []string) error {
	re := RepoEnvironment{}
	err := re.mustBeInANamespaceFolder()
	if err != nil {
		return err
	}

	engineValues, err := promptUserForRDSValues()
	if err != nil {
		return err
	}

	rdsTfFile, err := createRdsTfFile(engineValues)
	if err != nil {
		return err
	}

	fmt.Printf("RDS File generated in %s\n", rdsTfFile)
	color.Info.Tips("This template is using default values provided by your namespace information. Please review before raising PR")

	return nil
}

//------------------------------------------------------------------------------

func promptUserForRDSValues() (string, error) {
	q := userQuestion{
		description: heredoc.Doc(`
			 What RDS Engine you want to create?
			 Please enter "postgresql" or "mysql" or "mssql"
			 `),
		prompt:    "Engine",
		validator: new(rdsEngineValidator),
	}
	_ = q.getAnswer()
	return q.value, nil
}

func createRdsTfFile(engineValues string) (string, error) {
	// The rds "template" is actually an example file. Based on engineValues
	// fetch the relevant example file into the user's resources/ directory as `rds-<engine>.tf`
	rdsTemplateFile := rdsTemplateFilePrefix + "rds-" + engineValues + ".tf"
	rdsTfFile := rdsTfFilePrefix + "rds-" + engineValues + ".tf"

	if tfFileExists(rdsTfFile) {
		// generate a random string and append to the filename
		randomInt := strconv.Itoa(rand.Intn(10))
		rdsTfFile = rdsTfFilePrefix + "rds-" + engineValues + randomInt + ".tf"
	}

	err := CopyUrlToFile(rdsTemplateFile, rdsTfFile)
	if err != nil {
		return "", err
	}
	return rdsTfFile, nil
}

func tfFileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
