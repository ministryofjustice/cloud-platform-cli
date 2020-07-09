package environment

import (
	"errors"
	"regexp"
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

func (s *promptString) promptEmail() error {
	prompt := promptui.Prompt{
		Label:    s.label,
		Default:  s.defaultValue,
		Validate: validateEmailInput,
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

func promptSelectGithubTeam(t []string) (string, error) {

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "\U0001F336 {{ . | cyan }}",
		Inactive: "  {{ . | cyan }}",
		Selected: "\U0001F336 {{ . | red | cyan }}",
	}

	searcher := func(input string, index int) bool {
		team := t[index]
		name := strings.Replace(strings.ToLower(team), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:     "Github Team",
		Items:     t,
		Templates: templates,
		Size:      8,
		Searcher:  searcher,
	}

	i, _, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return t[i], nil
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

func validateEmailInput(input string) error {
	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if re.MatchString(input) == false {
		return errors.New("Please introduce a valid email address")
	}
	return nil
}
