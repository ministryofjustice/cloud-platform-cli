package environment

import (
	"bufio"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const cloudPlatformEnvRepo = "cloud-platform-environments"
const namespaceBaseFolder = "namespaces/live-1.cloud-platform.service.justice.gov.uk"

func outputFileWriter(fileName string) (*os.File, error) {
	f, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func downloadTemplate(url string) (string, error) {

	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	data, _ := ioutil.ReadAll(response.Body)
	content := string(data)

	return content, nil
}

// In the future we would like to point to one of our API servers
// The CLI shouldn't be getting this information by itself
func getGitHubTeams() ([]string, error) {
	var teams []string

	repo, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return nil, errors.New("You are outside cloud-platform-environment repo")
	}
	targetPath := strings.TrimSpace(string(repo)) + "/" + namespaceBaseFolder + "/"
	FullPath := strings.TrimSpace(string(repo))
	repoName := filepath.Base(FullPath)

	if repoName != "cloud-platform-environments" {
		return nil, errors.New("You are outside cloud-platform-environment repo")
	}

	re := regexp.MustCompile("github:(.*)\"")

	err = filepath.Walk(targetPath, func(path string, info os.FileInfo, err error) error {
		r, err := regexp.MatchString("01-rbac.yaml", info.Name())
		if err == nil && r {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				if strings.Contains(scanner.Text(), "github") {
					match := re.FindStringSubmatch(scanner.Text())
					if len(match) > 1 {
						teams = append(teams, match[1])
					}
				}
				if err := scanner.Err(); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	teamClean := removeDuplicates(teams)
	return teamClean, nil
}

func removeDuplicates(elements []string) []string {
	encountered := map[string]bool{}
	result := []string{}

	for v := range elements {
		if encountered[elements[v]] == true {
		} else {
			encountered[elements[v]] = true
			result = append(result, elements[v])
		}
	}
	return result
}
