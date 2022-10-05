package terraform

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// Redacted reads bytes of data for any sensitive strings and print REDACTED
func Redacted(w io.Writer, output string, redact bool) {
	re := regexp.MustCompile(`(?i)password|secret|token|key|https://hooks\.slack\.com|user|arn|ssh-rsa|clientid`)
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		if redact {
			if re.Match([]byte(scanner.Text())) {
				fmt.Fprintln(w, "REDACTED")
			} else {
				fmt.Fprintln(w, scanner.Text())
			}
		} else {
			fmt.Fprintln(w, scanner.Text())
		}
	}
}
