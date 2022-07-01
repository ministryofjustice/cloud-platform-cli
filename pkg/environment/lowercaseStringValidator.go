package environment

type lowercaseStringValidator struct{}

func (v *lowercaseStringValidator) isValid(s string) bool {
	r := new(regexValidator)
	r.regex = `^[a-z]+$`
	return r.isValid(s)
}
