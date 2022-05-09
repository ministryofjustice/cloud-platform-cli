package environment

type businessUnitValidator struct{}

func (v *businessUnitValidator) isValid(s string) bool {
	l := inListValidator{
		list: []string{
			"CICA",
			"HMCTS",
			"HMPPS",
			"LAA",
			"OPG",
			"Platforms",
			"HQ",
		},
	}
	return l.isValid(s)
}
