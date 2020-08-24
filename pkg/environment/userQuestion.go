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
	fmt.Println("")
	fmt.Println(q.description)

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
