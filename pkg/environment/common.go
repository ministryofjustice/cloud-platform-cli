package environment

import (
	"io/ioutil"
	"net/http"
	"os"
	"text/template"
)

const (
	cloudPlatformEnvRepo = "cloud-platform-environments"
	liveBaseDir          = "namespaces/live.cloud-platform.service.justice.gov.uk"
	devAlphaBaseDir      = "namespaces/dev-alpha.cloud-platform.service.justice.gov.uk"
	envTemplateLocation  = "https://raw.githubusercontent.com/ministryofjustice/cloud-platform-environments/main/namespace-resources-cli-template"
)

// namespaceBaseFolder is the base folder for the cloud-platform-environments repository.
// we set this as a global variable so it can be used to define the cluster directory later on.
var namespaceBaseFolder = liveBaseDir

type templateFromUrl struct {
	outputPath string
	content    string
	name       string
	url        string
}

func outputFileWriter(fileName string) (*os.File, error) {
	f, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func downloadTemplateContents(t []*templateFromUrl) error {
	for _, s := range t {
		content, err := downloadTemplate(s.url)
		if err != nil {
			return err
		}
		s.content = content
	}

	return nil
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

func directoryExists(path string) bool {
	if stat, err := os.Stat(path); err == nil && stat.IsDir() {
		return true
	} else {
		return false
	}
}

func CopyUrlToFile(url string, targetFilename string) error {
	str, err := downloadTemplate(url)
	if err != nil {
		return err
	}

	f, err := os.Create(targetFilename)
	if err != nil {
		return err
	}
	f.WriteString(str)
	f.Close()

	return nil
}

func createFilesFromTemplates(templates []*templateFromUrl, values Namespace) error {
	for _, i := range templates {
		t, err := template.New("").Parse(i.content)
		if err != nil {
			return err
		}

		f, err := os.Create(i.outputPath)
		if err != nil {
			return err
		}

		err = t.Execute(f, values)
		if err != nil {
			return err
		}
	}
	return nil
}
