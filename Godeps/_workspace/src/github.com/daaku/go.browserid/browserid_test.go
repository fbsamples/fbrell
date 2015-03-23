package browserid_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.browserid"
)

var bid = browserid.Cookie{
	Name:   "z",
	MaxAge: time.Hour * 24 * 365 * 10,
	Length: 16,
}

func TestHas(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("Error creating request: %s", err)
	}
	if bid.Has(req) {
		t.Fatalf("Error was not expecting request to have id")
	}
}

func TestGet(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("Error creating request: %s", err)
	}
	if bid.Has(req) {
		t.Fatalf("Error was not expecting request to have id")
	}
	w := httptest.NewRecorder()
	id1 := bid.Get(w, req)
	if w.Header().Get("Set-Cookie") == "" {
		t.Fatalf("Error was expecting a Set-Cookie header")
	}
	id2 := bid.Get(w, req)
	if id1 != id2 {
		t.Fatalf("Error got different ids: %s / %s", id1, id2)
	}
}
