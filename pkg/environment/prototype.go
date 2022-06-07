package environment

// Prototype represents a gov.uk prototype kit hosted on the cloud platform.
// Under the hood, this is a namespace and a github repository of the same
// name, connected via a github actions continuous deployment workflow and
// github actions secrets containing the namespace details, ecr &
// serviceaccount credentials.
type Prototype struct {
	Namespace         Namespace
	BasicAuthUsername string
	BasicAuthPassword string
}
