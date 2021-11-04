package bump

type BumpOptions struct {
	Version   string
	Module    string
	Namespace string
}

func ModuleVersion(opt BumpOptions) error {
	// walk environments file path
	// create a collection of all modules with that name
	// change the version of each module in the collection
	return nil
}
