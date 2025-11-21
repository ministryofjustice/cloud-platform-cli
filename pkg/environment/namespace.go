package environment

import (
	"fmt"
	"os"
	"strings"

	gogithub "github.com/google/go-github/github"
	"github.com/ministryofjustice/cloud-platform-cli/pkg/util"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
)

const NamespaceYamlFile = "00-namespace.yaml"

type Namespace struct {
	Application           string `yaml:"application"`
	BusinessUnit          string `yaml:"businessUnit"`
	Environment           string `yaml:"environment"`
	GithubTeam            string `yaml:"githubTeam"`
	InfrastructureSupport string `yaml:"infrastructureSupport"`
	IsProduction          string `yaml:"isProduction"`
	Namespace             string `yaml:"namespace"`
	Owner                 string `yaml:"owner"`
	OwnerEmail            string `yaml:"ownerEmail"`
	SlackChannel          string `yaml:"slackChannel"`
	SourceCode            string `yaml:"sourceCode"`
	ReviewAfter           string `yaml:"reviewAfter"`
	ServiceArea           string `yaml:"serviceArea"`
}

// This is a public function so that we can use it in our tests
func (ns *Namespace) ReadYaml() error {
	return ns.readYamlFile(NamespaceYamlFile)
}

func (ns *Namespace) readYamlFile(filename string) error {
	contents, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Failed to read namespace YAML file: %s", filename)
		return err
	}
	err = ns.parseYaml(contents)
	if err != nil {
		fmt.Printf("Failed to parse namespace YAML file: %s", filename)
		return err
	}
	return nil
}

func (ns *Namespace) parseYaml(yamlData []byte) error {
	type envNamespace struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
		Metadata   struct {
			Namespace string `yaml:"name"`
			Labels    struct {
				IsProduction string `yaml:"cloud-platform.justice.gov.uk/is-production"`
				Environment  string `yaml:"cloud-platform.justice.gov.uk/environment-name"`
			} `yaml:"labels"`
			Annotations struct {
				BusinessUnit string `yaml:"cloud-platform.justice.gov.uk/business-unit"`
				Application  string `yaml:"cloud-platform.justice.gov.uk/application"`
				Owner        string `yaml:"cloud-platform.justice.gov.uk/owner"`
				SourceCode   string `yaml:"cloud-platform.justice.gov.uk/source-code"`
				ReviewAfter  string `yaml:"cloud-platform.justice.gov.uk/review-after"`
			} `yaml:"annotations"`
		} `yaml:"metadata"`
	}

	t := envNamespace{}

	err := yaml.Unmarshal(yamlData, &t)
	if err != nil {
		fmt.Printf("Could not decode namespace YAML: %v", err)
		return err
	}

	ns.Application = t.Metadata.Annotations.Application
	ns.BusinessUnit = t.Metadata.Annotations.BusinessUnit
	ns.Environment = t.Metadata.Labels.Environment
	ns.IsProduction = t.Metadata.Labels.IsProduction
	ns.Namespace = t.Metadata.Namespace
	ns.Owner = t.Metadata.Annotations.Owner
	ns.OwnerEmail = strings.Split(t.Metadata.Annotations.Owner, ": ")[1]
	ns.SourceCode = t.Metadata.Annotations.SourceCode
	ns.ReviewAfter = t.Metadata.Annotations.ReviewAfter

	return nil
}

// nsCreateRawChangedFilesInPR get the list of changed files for a given PR. checks if the file is deleted and
// write the deleted file to the namespace folder
func (a *Apply) nsCreateRawChangedFilesInPR(cluster string, prNumber int) ([]string, error) {
	files, err := a.GithubClient.GetChangedFiles(prNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch list of changed files: %s", err)
	}

	namespaces, err := nsChangedInPR(files, a.Options.ClusterDir, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace for destroy from the PR: %s", err)
	}
	if len(namespaces) == 0 {
		fmt.Println("No namespace found in the PR for destroy")
		return nil, nil
	}
	canCreate, err := canCreateNamespaces(namespaces, cluster)
	if err != nil {
		return nil, fmt.Errorf("failed to create namespace for destroy: %s", err)
	}
	if !canCreate {
		fmt.Println("Cannot create namespace for destroy")
		return nil, nil
	}

	// Get the contents of the CommitFile from RawURL
	// https://developer.github.com/v3/repos/contents/#get-contents

	for _, file := range files {
		data, err := util.GetGithubRawContents(file.GetRawURL())
		if err != nil {
			return nil, fmt.Errorf("failed to get raw contents: %s", err)
		}
		// Create List with changed files
		if err := os.WriteFile(*file.Filename, data, 0o644); err != nil {
			return nil, fmt.Errorf("failed to write file list: %s", err)
		}
	}
	return namespaces, nil
}

// nsChangedInPR get the list of changed files for a given PR. checks if the namespaces exists in the given cluster
// folder and return the list of namespaces.
func nsChangedInPR(files []*gogithub.CommitFile, cluster string, isDeleted bool) ([]string, error) {
	var namespaceNames []string
	for _, file := range files {
		// check of the file is a deleted file
		if isDeleted && *file.Status != "removed" {
			fmt.Println("some of files are not marked for deletion: file", *file.Filename, "is not deleted")
			return nil, nil
		}

		// namespaces filepaths are assumed to come in
		// the format: namespaces/<cluster>.cloud-platform.service.justice.gov.uk/<namespaceName>
		s := strings.Split(*file.Filename, "/")
		// only get namespaces from the folder that belong to the given cluster and
		// ignore changes outside namespace directories
		if len(s) > 1 && s[1] == cluster {
			namespaceNames = append(namespaceNames, s[2])
		}
	}
	return util.DeduplicateList(namespaceNames), nil
}

func canCreateNamespaces(namespaces []string, cluster string) (bool, error) {
	wd, _ := os.Getwd()
	for _, ns := range namespaces {
		// make directory if it doesn't exist
		if _, err := os.Stat(wd + "/namespaces/" + cluster + "/" + ns); err != nil {
			err := os.Mkdir(wd+"/namespaces/"+cluster+"/"+ns, 0o755)
			if err != nil {
				return false, fmt.Errorf("error creating namespaces directory: %s", err)
			}
			err = os.Mkdir(wd+"/namespaces/"+cluster+"/"+ns+"/resources", 0o755)
			if err != nil {
				return false, fmt.Errorf("error creating resources directory: %s", err)
			}
		} else {
			fmt.Printf("namespace %s exists in the environments repo, skipping destroy", ns)
			return false, nil
		}
	}
	return true, nil
}

func isProductionNs(nsInPR string, namespaces []v1.Namespace) bool {
	for _, ns := range namespaces {
		is_prod := ns.Labels["cloud-platform.justice.gov.uk/is-production"] == "true"
		if ns.Name == nsInPR && is_prod {
			return true
		}
	}
	return false
}
