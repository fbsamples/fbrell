package mockoauth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestAuthorizeFormRendersConsentPage(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest("GET",
		Path+"authorize?response_type=code&client_id=123&redirect_uri=https://example.com/cb&scope=read,write&state=abc",
		nil)
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Authorize 123 for FBRell") {
		t.Fatal("consent page missing title")
	}
	if !strings.Contains(body, "Client ID: 123") {
		t.Fatal("consent page missing client_id")
	}
	if !strings.Contains(body, ">read</li>") {
		t.Fatal("consent page missing 'read' scope")
	}
	if !strings.Contains(body, ">write</li>") {
		t.Fatal("consent page missing 'write' scope")
	}
}

func TestAuthorizeFormNoScopes(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest("GET",
		Path+"authorize?response_type=code&client_id=456&redirect_uri=https://example.com/cb",
		nil)
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "No specific scopes requested") {
		t.Fatal("consent page missing no-scopes message")
	}
}

func TestAuthorizeFormMissingResponseType(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest("GET",
		Path+"authorize?client_id=123&redirect_uri=https://example.com/cb",
		nil)
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusBadRequest)
	}
	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["error"] != "invalid_request" {
		t.Fatalf("got error %q, want %q", resp["error"], "invalid_request")
	}
}

func TestAuthorizeFormInvalidResponseType(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest("GET",
		Path+"authorize?response_type=token&client_id=123&redirect_uri=https://example.com/cb",
		nil)
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusBadRequest)
	}
	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["error"] != "unsupported_response_type" {
		t.Fatalf("got error %q, want %q", resp["error"], "unsupported_response_type")
	}
}

func TestAuthorizeFormMissingClientID(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest("GET",
		Path+"authorize?response_type=code&redirect_uri=https://example.com/cb",
		nil)
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAuthorizeFormClientIDWithPipe(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest("GET",
		Path+"authorize?response_type=code&client_id=bad|id&redirect_uri=https://example.com/cb",
		nil)
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAuthorizeFormScopeWithPipe(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest("GET",
		Path+"authorize?response_type=code&client_id=123&redirect_uri=https://example.com/cb&scope=read|write",
		nil)
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAuthorizeFormMissingRedirectURI(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest("GET",
		Path+"authorize?response_type=code&client_id=123",
		nil)
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAuthorizeSubmitApprove(t *testing.T) {
	h := &Handler{}

	form := url.Values{}
	form.Set("client_id", "123")
	form.Set("redirect_uri", "https://example.com/cb")
	form.Set("state", "abc")
	form.Set("scope", "read,write")
	form.Set("action", "authorize")

	req := httptest.NewRequest("POST", Path+"authorize", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusFound {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusFound)
	}

	loc := w.Header().Get("Location")
	parsed, err := url.Parse(loc)
	if err != nil {
		t.Fatal(err)
	}
	code := parsed.Query().Get("code")
	if !strings.HasPrefix(code, "mock_code|123|read,write|valid|") {
		t.Fatalf("unexpected code: %s", code)
	}
	if parsed.Query().Get("state") != "abc" {
		t.Fatalf("state mismatch: got %q", parsed.Query().Get("state"))
	}
}


func TestAuthorizeSubmitRedirectURIWithQueryParams(t *testing.T) {
	h := &Handler{}

	form := url.Values{}
	form.Set("client_id", "123")
	form.Set("redirect_uri", "https://example.com/cb?existing=param&session=abc")
	form.Set("state", "xyz")
	form.Set("scope", "read,write")
	form.Set("action", "authorize")

	req := httptest.NewRequest("POST", Path+"authorize", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusFound {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusFound)
	}

	loc := w.Header().Get("Location")
	parsed, err := url.Parse(loc)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Query().Get("existing") != "param" {
		t.Fatalf("original query param lost: got %q", parsed.Query().Get("existing"))
	}
	if parsed.Query().Get("session") != "abc" {
		t.Fatalf("original query param lost: got %q", parsed.Query().Get("session"))
	}
	code := parsed.Query().Get("code")
	if !strings.HasPrefix(code, "mock_code|123|read,write|valid|") {
		t.Fatalf("unexpected code: %s", code)
	}
	if parsed.Query().Get("state") != "xyz" {
		t.Fatalf("state mismatch: got %q", parsed.Query().Get("state"))
	}
}

func TestAuthorizeFormRedirectURIWithFragment(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest("GET",
		Path+"authorize?response_type=code&client_id=123&redirect_uri=https://example.com/cb%23frag",
		nil)
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["error"] != "invalid_request" {
		t.Fatalf("got error %q, want %q", resp["error"], "invalid_request")
	}
}

func TestAuthorizeSubmitRedirectURIWithFragment(t *testing.T) {
	h := &Handler{}

	form := url.Values{}
	form.Set("client_id", "123")
	form.Set("redirect_uri", "https://example.com/cb#frag")
	form.Set("action", "authorize")

	req := httptest.NewRequest("POST", Path+"authorize", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["error"] != "invalid_request" {
		t.Fatalf("got error %q, want %q", resp["error"], "invalid_request")
	}
}

func TestUnknownEndpoint(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest("GET", Path+"unknown", nil)
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusNotFound)
	}
}

