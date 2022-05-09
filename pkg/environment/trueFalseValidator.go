package environment

type trueFalseValidator struct{}

func (v *trueFalseValidator) isValid(s string) bool {
	l := inListValidator{
		list: []string{"true", "false"},
	}
	return l.isValid(s)
}
