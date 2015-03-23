package trustforward_test

import (
	"net/http"
	"testing"

	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.trustforward"
)

var (
	allDisabled trustforward.Forwarded
	xEnabled    = trustforward.Forwarded{
		X: true,
	}
)

func TestHostTrustedButNotSet(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("Error creating request: %s", err)
	}
	actual := xEnabled.Host(req)
	const expected = "example.com"
	if actual != expected {
		t.Fatalf("Did not find expected host %s instead found %s", expected, actual)
	}
}

func TestHostTrustedAndNotSet(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("Error creating request: %s", err)
	}
	const expected = "foo.com"
	req.Header.Add("x-forwarded-host", expected)
	actual := xEnabled.Host(req)
	if actual != expected {
		t.Fatalf("Did not find expected host %s instead found %s", expected, actual)
	}
}

func TestHostNotTrustedSet(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("Error creating request: %s", err)
	}
	const expected = "example.com"
	req.Header.Add("x-forwarded-host", "foo.com")
	actual := allDisabled.Host(req)
	if actual != expected {
		t.Fatalf("Did not find expected host %s instead found %s", expected, actual)
	}
}
