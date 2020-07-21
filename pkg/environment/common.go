package environment

import (
	"io/ioutil"
	"net/http"
	"os"
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

func directoryExists(path string) bool {
	if stat, err := os.Stat(path); err == nil && stat.IsDir() {
		return true
	} else {
		return false
	}
}
