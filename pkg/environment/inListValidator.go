package environment

import (
	"fmt"
	"strings"
)

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
	fmt.Printf("Value must be in the list: %s\n", strings.Join(v.list, ", "))
	return false
}
