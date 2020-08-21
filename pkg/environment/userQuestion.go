package environment

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type userQuestion struct {
	description string
	prompt      string
	value       string
	validator   questionValidator
}

func (q *userQuestion) getAnswer() error {
	// TODO: Automatically format the description, restricting line-length
	fmt.Println(q.description)
	fmt.Print("\n")

	for {
		if q.validator.isValid(q.value) {
			break
		}

		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("%s: ", q.prompt)
		input, _ := reader.ReadString('\n')

		q.value = strings.TrimSpace(input)
	}
	return nil
}
