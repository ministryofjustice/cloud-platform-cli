package environment

type businessUnitValidator struct{}

func (v *businessUnitValidator) isValid(s string) bool {
	l := inListValidator{
		list: []string{
			"HQ",
			"HMPPS",
			"OPG",
			"LAA",
			"Central Digital",
			"Technology Services",
			"HMCTS",
			"CICA",
			"Platforms",
		},
	}
	return l.isValid(s)
}
