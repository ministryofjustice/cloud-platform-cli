package util

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/dlclark/regexp2"
)

type Repository struct {
	currentRepository string
	branch            string
}

type Date struct {
	First string
	Last  string
}

// set and return the name of the git repository which the current working
// directory is located within
func (re *Repository) Repository() (string, error) {
	// using re.repository here allows us to override this method in tests, so
	// that we can run tests regardless of the current working directory
	if re.currentRepository == "" {
		path, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
		if err != nil {
			return "", errors.New("current directory is not in a git repo working copy")
		}
		arr := strings.Split(string(path), "/")
		str := arr[len(arr)-1]
		re.currentRepository = strings.Trim(str, "\n")
	}

	return re.currentRepository, nil
}

// set and return the current branch of the git repository which is in the
// current working directory
func (re *Repository) GetBranch() (string, error) {
	// using re.getBranch here allows us to override this method in tests, so
	// that we can run tests regardless of the current working directory
	if re.branch == "" {
		// Fetch the git current branch
		branch, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
		if err != nil {
			fmt.Println("Cannot get the git branch. Check if it is a git repository")
			return "", err
		}
		re.branch = strings.Trim(string(branch), "\n")
	}
	return re.branch, nil
}

// It is assumed the current working directory is a git repo so ensure you check before calling this method
func GetLatestGitPull() error {
	reset := exec.Command("git", "reset", "--hard", "origin/main")
	pull := exec.Command("git", "pull")

	var out bytes.Buffer
	var stderr bytes.Buffer
	reset.Stdout = &out
	reset.Stderr = &stderr
	pull.Stdout = &out
	pull.Stderr = &stderr

	fmt.Println("Executing git reset")
	resetErr := reset.Run()
	if resetErr != nil {
		fmt.Println(fmt.Sprint(resetErr) + ": " + stderr.String())
		return resetErr
	}

	fmt.Println("Executing git pull")
	err := pull.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return err
	}

	return nil
}

// Redacted reads bytes of data for any sensitive strings and print REDACTED
// This functiion is used to prevent slack webhooks URLS from being output in pipeline logs
func Redacted(w io.Writer, output string, redact bool) {
	re := regexp2.MustCompile(`^.*https://hooks\.slack\.com.*$`, 0)
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		if redact {
			got, err := re.FindStringMatch(scanner.Text())
			if got != nil && err == nil {
				fmt.Fprintln(w, "REDACTED")
			} else {
				fmt.Fprintln(w, scanner.Text())
			}
		} else {
			fmt.Fprintln(w, scanner.Text())
		}
	}
}

// RedactedEnv read bytes of data for environment pipeline output, and where it
// encounters a resource block of type and name: random_id auth_token, replace sensitive
// contents of this block with the string REDACTED.

func RedactedEnv(w io.Writer, output string, redact bool) {
	// Regular expression to match elasticache auth_token: resource "random_id" "auth_token"
	re := regexp2.MustCompile(`(?=.*resource\s+"random_id"\s+"auth_token").*$`, 0)

	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {

		line := scanner.Text()

		if redact {

			match, err := re.MatchString(line)
			if err != nil {
				fmt.Println("Error", err)
			}

			// If current line matches regex, we have encountered an elasticache auth_token
			// resource block. Replace the fields contained within the block with the string
			// REDACTED
			if match {
				fmt.Fprintln(w, line)
				for scanner.Scan() {
					line = scanner.Text()
					// Check if the current line indicates the end of the resource block
					// and break out of the loop if it does
					if strings.Contains(strings.TrimSpace(line), "}") {
						break // Break out of the loop
					} else {
						continue // Continue until we reach the end of the resource block
					}
				}
				fmt.Fprintln(w, "REDACTED")
			} else {
				fmt.Fprintln(w, line)
			}
		} else {
			fmt.Fprintln(w, line)
		}
	}
}

func GetDatePastMinute(timestamp string, minutes int) (*Date, error) {
	d := &Date{}
	curTime, err := time.Parse("2006-01-02T15:04:05Z07:00", timestamp)
	if err != nil {
		return d, err
	}
	d.First = curTime.Truncate(time.Minute).Add(+time.Minute * time.Duration(1)).Format("2006-01-02T15:04:05")
	d.Last = curTime.Add(-time.Minute * time.Duration(minutes)).Format("2006-01-02T15:04:05")
	return d, nil
}

// DeduplicateList will simply take a slice of strings and
// return a deduplicated version.
func DeduplicateList(s []string) (list []string) {
	keys := make(map[string]bool)

	for _, entry := range s {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}

	return
}

func GetGithubRawContents(rawUrl string) ([]byte, error) {
	response, err := http.Get(rawUrl)

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("GetRawContents: Github File Raw Response is not 200 OK. Filename: %s, Response: %s", rawUrl, response.Status)
	}

	if err != nil {
		return nil, fmt.Errorf(" GetRawContents: Github File Raw Error: %s", err)
	}

	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("GetRawContents: Read Data Error: %s", err)
	}
	return data, nil
}

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
