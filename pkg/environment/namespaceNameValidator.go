package environment

type namespaceNameValidator struct{}

func (v *namespaceNameValidator) isValid(s string) bool {
	r := new(regexValidator)
	r.regex = `^[a-z0-9\-]+$`
	return r.isValid(s)
}
