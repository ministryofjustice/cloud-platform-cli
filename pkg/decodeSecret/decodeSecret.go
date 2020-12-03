package decodeSecret

import (
	"fmt"
)

type DecodeSecretOptions struct {
	Secret    string
	Namespace string
}

func DecodeSecret(opts *DecodeSecretOptions) error {
	fmt.Println("Hello from DecodeSecret")
	fmt.Println(opts.Namespace)
	fmt.Println(opts.Secret)
	return nil
}
