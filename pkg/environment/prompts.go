package environment

import (
	"errors"
	"net/url"
	"regexp"
	"strings"

	"github.com/manifoldco/promptui"
)

type promptTrueFalse struct {
	label        string
	defaultValue string
	value        string
}

type promptString struct {
	label        string
	defaultValue string
	value        string
	validation   string
}

//////////////
// PromptUI //
///////////////

func (s *promptString) promptString() error {
	prompt := promptui.Prompt{
		Label:   s.label,
		Default: s.defaultValue,
	}

	switch s.validation {
	case "email":
		prompt.Validate = validateEmailInput
	case "no-spaces-and-no-uppercase":
		prompt.Validate = validateWhiteSpacesAndUpperCase
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

func (s *promptTrueFalse) prompttrueFalse() error {
	prompt := promptui.Select{
		Label: s.label,
		Items: []string{"true", "false"},
	}
	_, result, err := prompt.Run()
	if err != nil {
		return err
	}
	s.value = result
	return nil
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

func validateWhiteSpacesAndUpperCase(input string) error {
	re := regexp.MustCompile(`\s`)
	re1 := regexp.MustCompile(`[A-Z]+`)

	if re.MatchString(input) == true {
		return errors.New("This input must consist of lower-case letters and dashes only (not whitespaces)")
	}
	if re1.MatchString(input) == true {
		return errors.New("This input must consist of lower-case letters only")
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
