package mockpartner

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseBearerTokenValid(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer mock_token|test_app|read,write")

	info, err := ParseBearerToken(req)
	if err != nil {
		t.Fatal(err)
	}
	if info.ClientID != "test_app" {
		t.Fatalf("got ClientID %q, want %q", info.ClientID, "test_app")
	}
	if len(info.Scopes) != 2 || info.Scopes[0] != "read" || info.Scopes[1] != "write" {
		t.Fatalf("got Scopes %v, want [read write]", info.Scopes)
	}
}

func TestParseBearerTokenNoScope(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer mock_token|test_app|noscope")

	info, err := ParseBearerToken(req)
	if err != nil {
		t.Fatal(err)
	}
	if info.ClientID != "test_app" {
		t.Fatalf("got ClientID %q, want %q", info.ClientID, "test_app")
	}
	if len(info.Scopes) != 0 {
		t.Fatalf("got Scopes %v, want empty", info.Scopes)
	}
}

func TestParseBearerTokenMissingHeader(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	_, err := ParseBearerToken(req)
	if err != ErrMissingAuth {
		t.Fatalf("got error %v, want %v", err, ErrMissingAuth)
	}
}

func TestParseBearerTokenNotBearer(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")

	_, err := ParseBearerToken(req)
	if err != ErrInvalidAuth {
		t.Fatalf("got error %v, want %v", err, ErrInvalidAuth)
	}
}

func TestParseBearerTokenNotMockToken(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer real_token_abc123")

	_, err := ParseBearerToken(req)
	if err != ErrInvalidToken {
		t.Fatalf("got error %v, want %v", err, ErrInvalidToken)
	}
}

func TestParseBearerTokenMalformed(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer mock_token|only_client_id")

	_, err := ParseBearerToken(req)
	if err != ErrInvalidToken {
		t.Fatalf("got error %v, want %v", err, ErrInvalidToken)
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	if err := WriteError(w, http.StatusUnauthorized, "invalid_token", "bad token"); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
