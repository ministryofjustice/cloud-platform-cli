package environment

type githubUrlValidator struct{}

func (v *githubUrlValidator) isValid(s string) bool {
	r := new(regexValidator)
	r.regex = `^https:\/\/github\.com\/[a-z\/\-_\.]+$`
	return r.isValid(s)
}
