package environment

type questionValidator interface {
	isValid(string) bool
}
