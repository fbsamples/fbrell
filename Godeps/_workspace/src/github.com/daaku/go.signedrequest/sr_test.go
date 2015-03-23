package signedrequest_test

import (
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.signedrequest"
	"testing"
)

type dummy struct {
	Value string `json:"0"`
}

func TestUnmarshal(t *testing.T) {
	data := []byte(
		"vlXgu64BQGFSQrY0ZcJBZASMvYvTHu9GQ0YM9rjPSso." +
			"eyJhbGdvcml0aG0iOiJITUFDLVNIQTI1NiIsIjAiOiJwYXlsb2FkIn0")
	secret := []byte("secret")
	out := new(dummy)
	err := signedrequest.Unmarshal(data, secret, out)
	if err != nil {
		t.Fatalf("Failed to Unmarshal: %s", err)
	}
	if out.Value != "payload" {
		t.Fatalf("Did not find expected value 'payload' instead found: %+v", out)
	}
}
