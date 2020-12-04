package decodeSecret

import "testing"

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
	err, actual := sd.processJson(jsn)
	if err != nil || actual != expected {
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
	err, actual := sd.processJson(jsn)
	if err != nil || actual != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s\n", expected, actual)
	}
}

func TestNoSuchSecret(t *testing.T) {
	sd := secretDecoder{}
	err, _ := sd.processJson("")
	if err == nil {
		t.Errorf("Expected an error")
	}
}
