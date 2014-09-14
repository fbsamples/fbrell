package trustforward_test

import (
	"flag"
	"github.com/daaku/go.trustforward"
	"net/http"
	"testing"
)

func TestHostTrustedButNotSet(t *testing.T) {
	flag.Set("trustforward", "1")
	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("Error creating request: %s", err)
	}
	actual := trustforward.Host(req)
	const expected = "example.com"
	if actual != expected {
		t.Fatalf("Did not find expected host %s instead found %s", expected, actual)
	}
}

func TestHostTrustedAndNotSet(t *testing.T) {
	flag.Set("trustforward", "1")
	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("Error creating request: %s", err)
	}
	const expected = "foo.com"
	req.Header.Add("x-forwarded-host", expected)
	actual := trustforward.Host(req)
	if actual != expected {
		t.Fatalf("Did not find expected host %s instead found %s", expected, actual)
	}
}

func TestHostNotTrustedSet(t *testing.T) {
	flag.Set("trustforward", "0")
	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("Error creating request: %s", err)
	}
	const expected = "example.com"
	req.Header.Add("x-forwarded-host", "foo.com")
	actual := trustforward.Host(req)
	if actual != expected {
		t.Fatalf("Did not find expected host %s instead found %s", expected, actual)
	}
}
