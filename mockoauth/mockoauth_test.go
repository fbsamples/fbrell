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

func TestTokenExchangeValid(t *testing.T) {
	h := &Handler{}
	code := BuildCode("789", "read,write", BehaviorValid)

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)

	req := httptest.NewRequest("POST", Path+"token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusOK)
	}

	var resp tokenResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.AccessToken != "mock_token|789|read,write" {
		t.Fatalf("got token %q, want %q", resp.AccessToken, "mock_token|789|read,write")
	}
	if resp.TokenType != "bearer" {
		t.Fatalf("got token_type %q, want %q", resp.TokenType, "bearer")
	}
	if resp.ExpiresIn != 3600 {
		t.Fatalf("got expires_in %d, want %d", resp.ExpiresIn, 3600)
	}
	if resp.Scope != "read,write" {
		t.Fatalf("got scope %q, want %q", resp.Scope, "read,write")
	}
}

func TestTokenExchangeNoScope(t *testing.T) {
	h := &Handler{}
	code := BuildCode("100", "", BehaviorValid)

	form := url.Values{}
	form.Set("code", code)

	req := httptest.NewRequest("POST", Path+"token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusOK)
	}

	var resp tokenResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.AccessToken != "mock_token|100|noscope" {
		t.Fatalf("got token %q, want %q", resp.AccessToken, "mock_token|100|noscope")
	}
	if resp.Scope != "" {
		t.Fatalf("got scope %q, want empty", resp.Scope)
	}
}

func TestTokenExchangeExpiredBehavior(t *testing.T) {
	h := &Handler{}
	code := BuildCode("123", "read", BehaviorExpired)

	form := url.Values{}
	form.Set("code", code)

	req := httptest.NewRequest("POST", Path+"token", strings.NewReader(form.Encode()))
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
	if resp["error"] != "invalid_grant" {
		t.Fatalf("got error %q, want %q", resp["error"], "invalid_grant")
	}
}

func TestTokenExchangeInvalidBehavior(t *testing.T) {
	h := &Handler{}
	code := BuildCode("123", "read", BehaviorInvalid)

	form := url.Values{}
	form.Set("code", code)

	req := httptest.NewRequest("POST", Path+"token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["error"] != "invalid_client" {
		t.Fatalf("got error %q, want %q", resp["error"], "invalid_client")
	}
}

func TestTokenExchangeMissingCode(t *testing.T) {
	h := &Handler{}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")

	req := httptest.NewRequest("POST", Path+"token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestTokenExchangeInvalidCode(t *testing.T) {
	h := &Handler{}

	form := url.Values{}
	form.Set("code", "not_a_valid_code")

	req := httptest.NewRequest("POST", Path+"token", strings.NewReader(form.Encode()))
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
	if resp["error"] != "invalid_grant" {
		t.Fatalf("got error %q, want %q", resp["error"], "invalid_grant")
	}
}

func TestTokenExchangeWrongGrantType(t *testing.T) {
	h := &Handler{}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("code", "mock_code|123|read|valid|12345")

	req := httptest.NewRequest("POST", Path+"token", strings.NewReader(form.Encode()))
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
	if resp["error"] != "unsupported_grant_type" {
		t.Fatalf("got error %q, want %q", resp["error"], "unsupported_grant_type")
	}
}

func TestTokenEndpointRejectsGet(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest("GET", Path+"token?code=mock_code|123|read|valid|12345", nil)
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestEndToEndConsentToToken(t *testing.T) {
	h := &Handler{}

	// Step 1: GET authorize — renders consent page
	req := httptest.NewRequest("GET",
		Path+"authorize?response_type=code&client_id=testapp&redirect_uri=https://partner.com/callback&scope=orders,products&state=xyz",
		nil)
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusOK {
		t.Fatalf("consent page got status %d, want %d", w.Code, http.StatusOK)
	}

	// Step 2: POST authorize — user clicks Confirm
	form := url.Values{}
	form.Set("client_id", "testapp")
	form.Set("redirect_uri", "https://partner.com/callback")
	form.Set("scope", "orders,products")
	form.Set("state", "xyz")
	form.Set("action", "authorize")

	req2 := httptest.NewRequest("POST", Path+"authorize", strings.NewReader(form.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w2 := httptest.NewRecorder()

	if err := h.Handle(w2, req2); err != nil {
		t.Fatal(err)
	}

	loc := w2.Header().Get("Location")
	parsed, err := url.Parse(loc)
	if err != nil {
		t.Fatal(err)
	}
	code := parsed.Query().Get("code")
	if code == "" {
		t.Fatal("no code in redirect")
	}

	// Step 3: Exchange code for token
	form3 := url.Values{}
	form3.Set("grant_type", "authorization_code")
	form3.Set("code", code)

	req3 := httptest.NewRequest("POST", Path+"token", strings.NewReader(form3.Encode()))
	req3.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w3 := httptest.NewRecorder()

	if err := h.Handle(w3, req3); err != nil {
		t.Fatal(err)
	}
	if w3.Code != http.StatusOK {
		t.Fatalf("token exchange got status %d, want %d", w3.Code, http.StatusOK)
	}

	var resp tokenResponse
	if err := json.Unmarshal(w3.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.AccessToken != "mock_token|testapp|orders,products" {
		t.Fatalf("got token %q, want %q", resp.AccessToken, "mock_token|testapp|orders,products")
	}
	if resp.Scope != "orders,products" {
		t.Fatalf("got scope %q, want %q", resp.Scope, "orders,products")
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

