package envs

import (
	"errors"
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
)

//////////////
// PromptUI //
//////////////

func promptString(name string) (string, error) {
	prompt := promptui.Prompt{
		Label:    name,
		Validate: validateEmptyInput,
	}

	return prompt.Run()
}

func promptyesNo(name string) (bool, error) {
	prompt := promptui.Select{
		Label: fmt.Sprintf("%s [Yes/No]", name),
		Items: []string{"Yes", "No"},
	}
	_, result, err := prompt.Run()
	if err != nil {
		return false, err
	}
	return result == "Yes", nil
}

func validateEmptyInput(input string) error {
	if len(strings.TrimSpace(input)) < 1 {
		return errors.New("This input must not be empty")
	}
	return nil
}
