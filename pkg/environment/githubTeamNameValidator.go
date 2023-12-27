package environment

type githubTeamNameValidator struct{}

func (v *githubTeamNameValidator) isValid(s string) bool {
	r := new(regexValidator)
	r.regex = `^[a-z0-9\-]+$`
	return r.isValid(s)
}
