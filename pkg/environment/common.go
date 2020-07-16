package environment

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

// all yaml and terraform templates will be pulled from URL endpoints below here
const templatesBaseUrl = "https://raw.githubusercontent.com/ministryofjustice/cloud-platform-environments/main/namespace-resources-cli-template"
const namespaceBaseFolder = "namespaces/live-1.cloud-platform.service.justice.gov.uk"

// metadataFromNamespace holds folder names (environments names) from
// cloud-platform-environments repository
type metadataFromNamespace struct {
	FileName        string
	Content         string
	isProduction    string
	environmentName string
	businessUnit    string
	application     string
	owner           string
	ownerEmail      string
	sourceCode      string
	namespace       string
	envRepoPath     string
}

func (s *metadataFromNamespace) getNamespaceMetadata() error {
	type envNamespace struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
		Metadata   struct {
			Name   string `yaml:"name"`
			Labels struct {
				CloudPlatformJusticeGovUkIsProduction    string `yaml:"cloud-platform.justice.gov.uk/is-production"`
				CloudPlatformJusticeGovUkEnvironmentName string `yaml:"cloud-platform.justice.gov.uk/environment-name"`
			} `yaml:"labels"`
			Annotations struct {
				CloudPlatformJusticeGovUkBusinessUnit string `yaml:"cloud-platform.justice.gov.uk/business-unit"`
				CloudPlatformJusticeGovUkApplication  string `yaml:"cloud-platform.justice.gov.uk/application"`
				CloudPlatformJusticeGovUkOwner        string `yaml:"cloud-platform.justice.gov.uk/owner"`
				CloudPlatformJusticeGovUkSourceCode   string `yaml:"cloud-platform.justice.gov.uk/source-code"`
			} `yaml:"annotations"`
		} `yaml:"metadata"`
	}

	t := envNamespace{}

	namespaceFile, err := ioutil.ReadFile(fmt.Sprintf("%s/namespaces/live-1.cloud-platform.service.justice.gov.uk/%s/00-namespace.yaml", s.envRepoPath, s.namespace))
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(namespaceFile, &t)
	if err != nil {
		log.Fatalf("Could not decode YAML (probably an error within 00-namespace.yaml file): %v", err)
		return err
	}

	s.isProduction = t.Metadata.Labels.CloudPlatformJusticeGovUkIsProduction
	s.businessUnit = t.Metadata.Annotations.CloudPlatformJusticeGovUkBusinessUnit
	s.owner = t.Metadata.Annotations.CloudPlatformJusticeGovUkOwner
	s.environmentName = t.Metadata.Labels.CloudPlatformJusticeGovUkEnvironmentName
	s.ownerEmail = strings.Split(t.Metadata.Annotations.CloudPlatformJusticeGovUkOwner, ": ")[1]
	s.application = t.Metadata.Annotations.CloudPlatformJusticeGovUkApplication
	s.sourceCode = t.Metadata.Annotations.CloudPlatformJusticeGovUkSourceCode
	s.namespace = t.Metadata.Name

	return nil
}

func (s *metadataFromNamespace) checkNamespaceExist() error {
	_, err := ioutil.ReadFile(fmt.Sprintf("%s/namespaces/live-1.cloud-platform.service.justice.gov.uk/%s/00-namespace.yaml", s.envRepoPath, s.namespace))
	if err != nil {
		return errors.New("You are in the wrong folder, go to your namespace folder")
	}
	return nil
}

func (s *metadataFromNamespace) checkPath() (bool, error) {
	path, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return false, errors.New("You are outside cloud-platform-environment repo")
	}
	FullPath := strings.TrimSpace(string(path))
	s.envRepoPath = FullPath
	repoName := filepath.Base(FullPath)

	if repoName != "cloud-platform-environments" {
		return false, errors.New("You are outside cloud-platform-environment repo")
	}

	return true, nil
}

func (s *metadataFromNamespace) getNamespaceFromPath() error {
	path, err := os.Getwd()
	if err != nil {
		return err
	}

	parts := strings.Split(path, "/")
	s.namespace = parts[len(parts)-1]

	return nil
}

func outputFileWriter(fileName string) (*os.File, error) {
	EnvRepoPath, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return nil, err
	}

	fullPath := fmt.Sprintf("%s/%s", strings.TrimSpace(string(EnvRepoPath)), fileName)

	f, err := os.Create(fullPath)
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
	targetPath := strings.TrimSpace(string(repo)) + "/namespaces/live-1.cloud-platform.service.justice.gov.uk/"
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
