package decodeSecret

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os/exec"
)

type DecodeSecretOptions struct {
	Secret    string
	Namespace string
}

func DecodeSecret(opts *DecodeSecretOptions) error {
	jsn := retrieveSecret(opts.Namespace, opts.Secret)

	err, str := processJson(jsn)
	if err != nil {
		fmt.Println("Error: ", err)
		return err
	}

	fmt.Printf(str)
	return nil
}

func processJson(jsn string) (error, string) {
	var result map[string]interface{}
	json.Unmarshal([]byte(jsn), &result)

	data := result["data"].(map[string]interface{})

	err := decodeKeys(data)
	if err != nil {
		return err, ""
	}

	err, str := formatJson(result)
	if err != nil {
		return err, ""
	}

	return nil, str
}

func retrieveSecret(namespace, secret string) string {
	cmd := exec.Command("kubectl", "--namespace", namespace, "get", "secret", secret, "-o", "json")

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		fmt.Println("Error: ", err)
		return ""
	}

	return out.String()
}

func decodeKeys(data map[string]interface{}) error {
	for k, v := range data {
		switch v.(type) {
		case string:
			data[k] = base64decode(v)
		default:
			return fmt.Errorf("Expected key %s of secret to be a string, but it wasn't\n", k)
		}
	}
	return nil
}

func base64decode(i interface{}) string {
	str, e := base64.StdEncoding.DecodeString(i.(string))
	if e != nil {
		fmt.Println(e)
		return "ERROR: base64 decode failed"
	}
	return fmt.Sprintf("%s", str)
}

func formatJson(result map[string]interface{}) (error, string) {
	str, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		return err, ""
	}

	return nil, fmt.Sprintf("%s\n", str)
}
