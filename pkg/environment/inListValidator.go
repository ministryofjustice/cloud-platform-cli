package environment

import "fmt"

type inListValidator struct {
	list []string
}

func (v *inListValidator) isValid(s string) bool {
	for _, i := range v.list {
		if s == i {
			return true
		}
	}
	// TODO: output the list
	fmt.Println("Value must be in the list")
	return false
}
