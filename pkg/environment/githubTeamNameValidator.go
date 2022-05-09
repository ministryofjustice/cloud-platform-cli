package environment

type githubTeamNameValidator struct{}

func (v *githubTeamNameValidator) isValid(s string) bool {
	r := new(regexValidator)
	r.regex = `^[a-z\-]+$`
	return r.isValid(s)
}
