package enviroment

import (
	"errors"
	"strings"

	"github.com/manifoldco/promptui"
)

type promptYesNo struct {
	label        string
	defaultValue int
	value        bool
}

type promptString struct {
	label        string
	defaultValue string
	value        string
}

//////////////
// PromptUI //
//////////////

func (s *promptString) promptString() error {
	prompt := promptui.Prompt{
		Label:    s.label,
		Default:  s.defaultValue,
		Validate: validateEmptyInput,
	}

	result, err := prompt.Run()
	if err != nil {
		return err
	}

	s.value = result
	return nil
}

func (s *promptYesNo) promptyesNo() error {
	prompt := promptui.Select{
		Label:     s.label,
		Items:     []string{"Yes", "No"},
		CursorPos: s.defaultValue,
	}
	_, result, err := prompt.Run()
	if err != nil {
		return err
	}
	s.value = result == "Yes"
	return nil
}

func promptSelectNamespaces(e *[]NamespacesFromGH) (string, error) {

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "\U0001F336 {{ .Name | cyan }}",
		Inactive: "  {{ .Name | cyan }}",
		Selected: "\U0001F336 {{ .Name | red | cyan }}",
	}

	searcher := func(input string, index int) bool {
		environment := (*e)[index]
		name := strings.Replace(strings.ToLower(environment.Name), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:     "Live Environments",
		Items:     *e,
		Templates: templates,
		Size:      8,
		Searcher:  searcher,
	}

	i, _, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return (*e)[i].Name, nil
}

/////////////////
// Validations //
/////////////////

func validateEmptyInput(input string) error {
	if len(strings.TrimSpace(input)) < 1 {
		return errors.New("This input must not be empty")
	}
	return nil
}
