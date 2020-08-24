package environment

type slackChannelValidator struct {
}

func (v *slackChannelValidator) isValid(s string) bool {
	r := new(regexValidator)
	r.regex = `^[a-z\-]+$`
	return r.isValid(s)
}
