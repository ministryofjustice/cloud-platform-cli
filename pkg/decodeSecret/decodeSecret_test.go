package decodeSecret

import (
	"strings"
	"testing"
)

func TestDecodeSecret(t *testing.T) {
	jsn := `{ "data": { "key1": "d2liYmxl", "key2": "d29iYmxl" } }`

	expected := `{
    "data": {
        "key1": "wibble",
        "key2": "wobble"
    }
}
`

	sd := secretDecoder{}
	actual, err := sd.processJson(jsn)
	if err != nil || actual != strings.TrimSpace(expected) {
		t.Errorf("Expected:\n%s\nGot:\n%s\n", expected, actual)
	}
}

func TestBadBase64(t *testing.T) {
	jsn := `{ "data": { "key1": "1", "key2": "2" } }`

	expected := `{
    "data": {
        "key1": "ERROR: base64 decode failed",
        "key2": "ERROR: base64 decode failed"
    }
}
`
	sd := secretDecoder{}
	actual, err := sd.processJson(jsn)
	if err != nil || actual != strings.TrimSpace(expected) {
		t.Errorf("Expected:\n%s\nGot:\n%s\n", expected, actual)
	}
}

func TestNoSuchSecret(t *testing.T) {
	sd := secretDecoder{}
	_, err := sd.processJson("")
	if err == nil {
		t.Errorf("Expected an error")
	}
}

func TestRecordsAwsSecrets(t *testing.T) {
	// jsn := `{ "data": { "access_key_id": "myaccesskey", "secret_access_key": "mysecretkey" } }`
	jsn := `{ "data": { "access_key_id": "bXlhY2Nlc3NrZXk=", "secret_access_key": "bXlzZWNyZXRrZXk=" } }`
	sd := secretDecoder{}
	_, err := sd.processJson(jsn)
	if err != nil || sd.AccessKeyID != "myaccesskey" {
		t.Errorf("Expected:\n%s\nGot:\n%s\n", "myaccesskey", sd.AccessKeyID)
	}
	if err != nil || sd.SecretAccessKey != "mysecretkey" {
		t.Errorf("Expected:\n%s\nGot:\n%s\n", "mysecretkey", sd.SecretAccessKey)
	}
}
