package decodeSecret

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
)

type secretDecoder struct {
	AccessKeyID     string
	SecretAccessKey string
}

type DecodeSecretOptions struct {
	Secret         string
	Namespace      string
	ExportAwsCreds bool
	// Raw allows you to output the json without formatting it.
	// Formatting the json means we can't print unicode characters (that might exist in secrets).
	Raw bool
}

func DecodeSecret(opts *DecodeSecretOptions) error {
	jsn := retrieveSecret(opts.Namespace, opts.Secret)

	sd := secretDecoder{}

	str, err := sd.processJson(jsn, opts.Raw)
	if err != nil {
		return err
	}

	if opts.ExportAwsCreds {
		fmt.Println("export AWS_REGION=\"eu-west-2\"")
		fmt.Println("export AWS_ACCESS_KEY_ID=\"" + sd.AccessKeyID + "\"")
		fmt.Println("export AWS_SECRET_ACCESS_KEY=\"" + sd.SecretAccessKey + "\"")
	} else {
		fmt.Println(str)
	}
	return nil
}

func retrieveSecret(namespace, secret string) string {

	// declare cmd variable
	var cmd *exec.Cmd

	if namespace == "" {
		cmd = exec.Command("kubectl", "get", "secret", secret, "-o", "json")
	} else {
		cmd = exec.Command("kubectl", "--namespace", namespace, "get", "secret", secret, "-o", "json")
	}

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		fmt.Println("Error: ", err)
		return ""
	}

	return out.String()
}

func (sd *secretDecoder) processJson(jsn string, rawPrint bool) (string, error) {
	if jsn == "" {
		return "", errors.New("failed to retrieve secret from namespace")
	}

	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsn), &result)
	if err != nil {
		return "", err
	}

	data := result["data"].(map[string]interface{})

	err = decodeKeys(data)
	if err != nil {
		return "", err
	}

	sd.stashAwsCredentials(data)

	// If the user selected to output the raw JSON, do so
	var rawStr string
	if rawPrint {
		for k, v := range data {
			rawStr += fmt.Sprintf("%s: %s\n", k, v)
		}
		if rawStr == "" {
			return "", errors.New("unable to return the raw JSON, either the secret is empty or the secret is not in the correct format")
		}

		return rawStr, nil
	}

	str, err := formatJson(result)
	if err != nil {
		return "", err
	}

	return str, nil
}

// Stash AWS creds, if present, in case we need to output commands to set them
// as shell variables
func (sd *secretDecoder) stashAwsCredentials(data map[string]interface{}) {
	if val, ok := data["access_key_id"].(string); ok {
		sd.AccessKeyID = val
	}

	if val, ok := data["secret_access_key"].(string); ok {
		sd.SecretAccessKey = val
	}
}

func decodeKeys(data map[string]interface{}) error {
	for k, v := range data {
		switch v.(type) {
		case string:
			data[k] = base64decode(v)
		default:
			return fmt.Errorf("expected key %s of secret to be a string, but it wasn't", k)
		}
	}
	return nil
}

func base64decode(i interface{}) string {
	str, e := base64.StdEncoding.DecodeString(i.(string))
	if e != nil {
		return "ERROR: base64 decode failed"
	}
	return string(str)
}

func formatJson(result map[string]interface{}) (string, error) {
	str, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		return "", err
	}

	return string(str), nil
}
