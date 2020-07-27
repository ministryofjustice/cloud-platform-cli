package kubecfg

import (
	"reflect"
	"testing"
)

func eq(t *testing.T, got interface{}, expected interface{}) {
	t.Helper()
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected: %v, got: %v", expected, got)
	}
}

func TestGetTokenClaims(t *testing.T) {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkFsZWphbmRybyBHYXJyaWRvIE1vdGEiLCJlbWFpbCI6ImFsZWphbmRyby5nYXJyaWRvQGRpZ2l0YWwuanVzdGljZS5nb3YudWsiLCJodHRwczovL2s4cy5pbnRlZ3JhdGlvbi5kc2QuaW8vZ3JvdXBzIjpbbnVsbCwiZ2l0aHViOndlYm9wcyIsbnVsbF0sIm5pY2tuYW1lIjoibW9nYWFsIiwiaWF0IjoxNTE2MjM5MDIyfQ._5Vkfh1lDWQvV9xIcbpnMFWOnkJBzCcp1X7lYcEG9sE"

	claims, _ := getTokenClaims(token)
	eq(t, (*claims)["name"], "Alejandro Garrido Mota")
	eq(t, (*claims)["nickname"], "mogaal")
	eq(t, (*claims)["email"], "alejandro.garrido@digital.justice.gov.uk")
}

func TestGetToken(t *testing.T) {
	token, _ := getToken("./fixtures/kubeconfig.yml")
	if token[0] != "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkFsZWphbmRybyBHYXJyaWRvIE1vdGEiLCJlbWFpbCI6ImFsZWphbmRyby5nYXJyaWRvQGRpZ2l0YWwuanVzdGljZS5nb3YudWsiLCJodHRwczovL2s4cy5pbnRlZ3JhdGlvbi5kc2QuaW8vZ3JvdXBzIjpbbnVsbCwiZ2l0aHViOndlYm9wcyIsbnVsbF0sIm5pY2tuYW1lIjoibW9nYWFsIiwiaWF0IjoxNTE2MjM5MDIyfQ._5Vkfh1lDWQvV9xIcbpnMFWOnkJBzCcp1X7lYcEG9sE" {
		t.Errorf("Unexpected token, error reading the token from fixtures: %s", token[0])
	}
}
