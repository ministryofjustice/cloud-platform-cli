package environment

type teamEmailValidator struct {
}

func (v *teamEmailValidator) isValid(s string) bool {
	r := new(regexValidator)
	r.regex = `^[a-z\-\.]+\@[a-z\-\.]+$`
	return r.isValid(s)
}
