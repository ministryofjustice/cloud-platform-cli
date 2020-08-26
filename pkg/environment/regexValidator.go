package environment

import (
	"fmt"
	"regexp"
)

type regexValidator struct {
	regex string
}

func (v *regexValidator) isValid(s string) bool {
	matched, _ := regexp.MatchString(v.regex, s)
	if !matched {
		fmt.Printf("Value must match: %s\n", v.regex)
		return false
	}
	return true
}
