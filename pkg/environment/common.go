package enviroment

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	b64 "encoding/base64"

	"gopkg.in/yaml.v2"
)

// NamespacesFromGH holds namespaces names (folders) from
// cloud-platform-environments repository
type NamespacesFromGH struct {
	Name string `json:"name"`
}

// MetaDataFromGH holds folder names (environments names) from
// cloud-platform-environments repository
type MetaDataFromGH struct {
	FileName        string `json:"name"`
	Content         string `json:"content"`
	isProduction    string
	environmentName string
	businessUnit    string
	application     string
	owner           string
	ownerEmail      string
	sourceCode      string
	namespace       string
}

// GetNamespacesFromGH returns the environments names from Cloud Platform Environments
// repository (in Github)
func GetNamespacesFromGH() (*[]NamespacesFromGH, error) {
	var n []NamespacesFromGH

	response, err := http.Get("https://api.github.com/repos/ministryofjustice/cloud-platform-environments/contents/namespaces/live-1.cloud-platform.service.justice.gov.uk")
	if err != nil {
		return nil, err
	}

	data, _ := ioutil.ReadAll(response.Body)
	err = json.Unmarshal(data, &n)
	return &n, nil
}

// GetEnvironmentsMetadataFromGH returns the metadata about an environment from
// Cloud Platform Environments repository (in Github)
func (s *MetaDataFromGH) GetEnvironmentsMetadataFromGH() error {
	url := fmt.Sprintf("https://api.github.com/repos/ministryofjustice/cloud-platform-environments/contents/namespaces/live-1.cloud-platform.service.justice.gov.uk/%s/00-namespace.yaml", s.namespace)

	response, err := http.Get(url)
	if err != nil {
		return err
	}

	data, _ := ioutil.ReadAll(response.Body)
	err = json.Unmarshal(data, &s)
	if err != nil {
		log.Fatalf("Could not decode JSON from Github (containing the metadata): %v", err)
		return err
	}

	// I know about bellow! But is the safest way how to do it in Go
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

	sDec, _ := b64.StdEncoding.DecodeString(s.Content)
	s.Content = string(sDec)

	err = yaml.Unmarshal([]byte(s.Content), &t)
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

// downloadTemplate returns a template file from an URL
func downloadTemplate(url string) (string, error) {

	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	data, _ := ioutil.ReadAll(response.Body)
	content := string(data)

	return content, nil
}

func validPath() (bool, error) {
	path, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return false, err
	}

	FullPath := strings.TrimSpace(string(path))
	repoName := filepath.Base(FullPath)

	if repoName != "cloud-platform-environments" {
		outsidePath := promptYesNo{
			label:        "WARNING: You are outside cloud-platform-environment repo. If you decide to continue the template is going to be rendered on screen?",
			defaultValue: 0,
		}
		err = outsidePath.promptyesNo()
		if err != nil {
			return false, err
		}

		if outsidePath.value == false {
			os.Exit(0)
		}
	}

	return false, nil
}
