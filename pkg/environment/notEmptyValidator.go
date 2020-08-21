package environment

import "fmt"

type notEmptyValidator struct {
}

func (v *notEmptyValidator) isValid(s string) bool {
	if s == "" {
		fmt.Println("A value is required")
		return false
	}
	return true
}
