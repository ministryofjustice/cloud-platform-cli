package environment

import (
	"bufio"
	"fmt"
	"os"
)

type userQuestion struct {
	description string
	prompt      string
	value       string
}

func (q *userQuestion) getAnswer() error {
	// TODO: Automatically format the description, restricting line-length
	fmt.Println(q.description)
	fmt.Print("\n")
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s: ", q.prompt)
	input, _ := reader.ReadString('\n')

	// validation loop here

	q.value = input
	return nil
}
