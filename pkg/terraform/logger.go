package terraform

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

// CmdOutput has the Stout and Stderr
type CmdOutput struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func (o *CmdOutput) redacted() {
	re := regexp.MustCompile(`(?i)password|secret|token|key|https://hooks.slack.com`)
	scanner := bufio.NewScanner(strings.NewReader(o.Stdout))

	for scanner.Scan() {
		if re.Match([]byte(scanner.Text())) {
			fmt.Println("REDACTED")
		} else {
			fmt.Println(scanner.Text())
		}
	}
}
