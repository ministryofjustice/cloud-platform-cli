package environment

type rdsEngineValidator struct{}

func (r *rdsEngineValidator) isValid(s string) bool {
	l := inListValidator{
		list: []string{
			"postgresql",
			"mysql",
			"mssql",
		},
	}
	return l.isValid(s)
}
