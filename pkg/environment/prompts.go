package environment

import (
	"errors"
	"net/url"
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
	validation   string
}

//////////////
// PromptUI //
//////////////

func (s *promptString) promptString() error {
	prompt := promptui.Prompt{
		Label:   s.label,
		Default: s.defaultValue,
	}

	switch s.validation {
	case "email":
		prompt.Validate = validateEmailInput
	case "no-spaces":
		prompt.Validate = validateWhiteSpaces
	case "url":
		prompt.Validate = validateURL
	default:
		prompt.Validate = validateEmptyInput
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

func validateWhiteSpaces(input string) error {
	re := regexp.MustCompile(`\s`)
	if re.MatchString(input) == true {
		return errors.New("This input must consist of lower-case letters and dashes only (not whitespaces)")
	}
	if len(strings.TrimSpace(input)) < 1 {
		return errors.New("This input must not be empty")
	}
	return nil
}

func validateURL(input string) error {
	u, err := url.Parse(input)
	if err != nil {
		return errors.New("Not valid URL. Please introduce a valid URL")
	} else if u.Scheme == "" || u.Host == "" {
		return errors.New("Not valid URL. Valid URL must be an absolute URL")
	} else if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("Not valid URL. URLs should start with http or https")
	}
	return nil
}
