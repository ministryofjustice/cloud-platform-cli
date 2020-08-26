package environment

type namespaceNameValidator struct {
}

func (v *namespaceNameValidator) isValid(s string) bool {
	r := new(regexValidator)
	r.regex = `^[a-z\-]+$`
	return r.isValid(s)
}
