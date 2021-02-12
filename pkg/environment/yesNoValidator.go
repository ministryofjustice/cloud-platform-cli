package environment

type yesNoValidator struct {
}

func (v *yesNoValidator) isValid(s string) bool {
	l := inListValidator{
		list: []string{"yes", "no"},
	}
	return l.isValid(s)
}
