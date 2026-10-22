package web

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOAuthCallbackIsCredentialIndependent(t *testing.T) {
	for _, query := range []string{
		"",
		"?code=fake_secret_code&state=fake_secret_state",
		"?access_token=fake_secret_token&signed_request=fake_signed_request",
	} {
		t.Run(query, func(t *testing.T) {
			var logs bytes.Buffer
			// No mux or Rell context: the callback must not enter either middleware.
			h := &Handler{Logger: log.New(&logs, "", 0)}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest(http.MethodGet,
				"https://www.fbrell.com"+oauthCallbackPath+query, nil))
			if w.Code != http.StatusOK || w.Body.String() != oauthCallbackNotice {
				t.Fatalf("got status %d and body %q, want 200 and the fixed callback notice", w.Code, w.Body.String())
			}
			if logs.Len() != 0 {
				t.Fatalf("callback wrote to the application error logger: %q", logs.String())
			}
			if w.Header().Get("Location") != "" {
				t.Fatal("callback must not introduce an extra redirect")
			}
			for key, expected := range map[string]string{
				"Content-Type":            "text/plain; charset=utf-8",
				"Cache-Control":           "no-store",
				"Referrer-Policy":         "no-referrer",
				"X-Content-Type-Options":  "nosniff",
				"Content-Security-Policy": "default-src 'none'; frame-ancestors 'none'",
			} {
				if value := w.Header().Get(key); value != expected {
					t.Fatalf("got %s %q, want %q", key, value, expected)
				}
			}
		})
	}
}

func TestOAuthCallbackHeadHasNoBody(t *testing.T) {
	w := httptest.NewRecorder()
	(&Handler{}).ServeHTTP(w, httptest.NewRequest(http.MethodHead, oauthCallbackPath, nil))
	if w.Code != http.StatusOK || w.Body.Len() != 0 {
		t.Fatalf("got status %d and body %q, want 200 with no body", w.Code, w.Body.String())
	}
}

func TestOAuthCallbackRejectsPostWithoutReadingCredentials(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, oauthCallbackPath,
		strings.NewReader("code=fake_secret_code&state=fake_secret_state"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	(&Handler{}).ServeHTTP(w, r)
	if w.Code != http.StatusMethodNotAllowed || w.Header().Get("Allow") != "GET, HEAD" {
		t.Fatalf("got status %d and Allow %q, want 405 and GET, HEAD", w.Code, w.Header().Get("Allow"))
	}
	if strings.Contains(w.Body.String(), "fake_secret") {
		t.Fatal("callback echoed posted authorization data")
	}
}
