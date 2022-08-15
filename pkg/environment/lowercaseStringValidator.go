package environment

type lowercaseStringValidator struct{}

func (v *lowercaseStringValidator) isValid(s string) bool {
	r := new(regexValidator)
	r.regex = `^\w+-\w+$`
	return r.isValid(s)
}
