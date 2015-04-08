package xsrf_test

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/daaku/rell/internal/github.com/daaku/go.browserid"
	"github.com/daaku/rell/internal/github.com/daaku/go.xsrf"
)

const (
	serverURL = "http://example.com/"
	bitUno    = "bitUno"
)

var provider = xsrf.ProviderFlag("default-provider")

func init() {
	provider.BrowserID = browserid.CookieFlag("browserid")
}

func TestToken(t *testing.T) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", serverURL, nil)
	if err != nil {
		t.Fatalf("Unexpected error creating new request: %s", err)
	}
	token1 := provider.Token(w, req)
	if token1 == "" {
		t.Fatalf("Was expecting non empty token1.")
	}
	token2 := provider.Token(w, req, bitUno)
	if token2 == "" {
		t.Fatalf("Was expecting non empty token2.")
	}
	if token1 == token2 {
		t.Fatalf("Was expecting different tokens.")
	}
	if !provider.Validate(token1, w, req) {
		t.Fatalf("Failed to validate token1.")
	}
	if !provider.Validate(token2, w, req, bitUno) {
		t.Fatalf("Failed to validate token2.")
	}
	if provider.Validate("", w, req) {
		t.Fatalf("Empty token should not be valid.")
	}
	if provider.Validate(token1, w, req, "foo") {
		t.Fatalf("Token should not be valid for foo.")
	}
}

func TestBadButWellEncoded(t *testing.T) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", serverURL, nil)
	if err != nil {
		t.Fatalf("Unexpected error creating new request: %s", err)
	}
	encoded := base64.URLEncoding.EncodeToString([]byte("foo"))
	if provider.Validate(encoded, w, req, "foo") {
		t.Fatal("Token should not be valid.")
	}
}
