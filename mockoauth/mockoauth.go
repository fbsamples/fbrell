/**
 * Copyright (c) 2014-present, Facebook, Inc. All rights reserved.
 *
 * You are hereby granted a non-exclusive, worldwide, royalty-free license to use,
 * copy, modify, and distribute this software in source code or binary form for use
 * in connection with the web services and APIs provided by Facebook.
 *
 * As with any software that integrates with the Facebook platform, your use of
 * this software is subject to the Facebook Developer Principles and Policies
 * [http://developers.facebook.com/policy/]. This copyright notice shall be
 * included in all copies or substantial portions of the software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

// Package mockoauth implements a mock OAuth 2.0 authorization server for testing
// Meta's 3P partner integration flows. Unlike the oauth package (where FBRell is
// a client of Facebook's OAuth), here FBRell acts as the partner's OAuth provider.
//
// Two endpoints implement the authorization code grant (RFC 6749 §4.1):
//   - GET/POST /mock-oauth/authorize — renders a consent screen and issues an
//     authorization code on user approval.
//   - POST /mock-oauth/token — exchanges an authorization code for an access token.
//
// Codes and tokens are non-cryptographic, human-readable strings encoding the
// client ID, granted scopes, and configurable behavior (valid/expired/invalid).
//
// Client authentication on the token endpoint follows RFC 6749 §2.3.1. Both
// HTTP Basic Auth and form-body credentials are accepted. The expected
// client_secret for any client_id is deterministic and stateless: callers
// configure their OAuth client with `mock_secret_<client_id>`. This avoids
// the need for server-side credential storage while still exercising the
// credential round-trip behavior of real OAuth clients.
package mockoauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/daaku/go.h"
)

const (
	Path = "/mock-oauth/"

	// secretPrefix is the deterministic prefix used to derive the expected
	// client_secret for a given client_id. See ValidateClient.
	secretPrefix = "mock_secret_"
)

var (
	errMissingClientID          = errors.New("mock-oauth: missing client_id parameter")
	errMissingClientSecret      = errors.New("mock-oauth: missing client_secret parameter")
	errInvalidClientCredentials = errors.New("mock-oauth: client authentication failed")
	errClientIDMismatch         = errors.New("mock-oauth: client_id does not match the authorization code")
	errCredentialsConflict      = errors.New("mock-oauth: conflicting credentials in Basic auth header and request body")
	errMissingRedirectURI       = errors.New("mock-oauth: missing redirect_uri parameter")
	errMissingResponseType      = errors.New("mock-oauth: missing response_type parameter")
	errInvalidResponseType      = errors.New("mock-oauth: response_type must be 'code'")
	errRedirectURIFragment      = errors.New("mock-oauth: redirect_uri must not contain a fragment")
	errMissingCode              = errors.New("mock-oauth: missing code parameter")
	errInvalidCode              = errors.New("mock-oauth: invalid or malformed authorization code")
	errExpiredCode              = errors.New("mock-oauth: authorization code has expired")
	errInvalidGrantType         = errors.New("mock-oauth: grant_type must be authorization_code")
	errInvalidClientIDChar      = errors.New("mock-oauth: client_id must not contain '|'")
	errInvalidScopeChar         = errors.New("mock-oauth: scope values must not contain '|'")
)

// ValidateClient reports whether the given client_secret is correct for the
// given client_id under the mock's deterministic, stateless secret format.
// Configure your OAuth client (e.g. MC3P Authoring Tool) with the secret
// `mock_secret_<client_id>` to authenticate against this mock.
func ValidateClient(clientID, secret string) bool {
	return secret == secretPrefix+clientID
}

// TokenBehavior controls what kind of token the /token endpoint returns.
type TokenBehavior string

const (
	BehaviorValid   TokenBehavior = "valid"
	BehaviorExpired TokenBehavior = "expired"
	BehaviorInvalid TokenBehavior = "invalid"
)

// Handler serves mock OAuth endpoints for testing OAuth flows.
type Handler struct{}

// Handle routes requests to the appropriate mock OAuth endpoint.
func (a *Handler) Handle(w http.ResponseWriter, r *http.Request) error {
	switch r.URL.Path {
	case Path + "authorize":
		if r.Method == http.MethodPost {
			return a.AuthorizeSubmit(w, r)
		}
		return a.AuthorizeForm(w, r)
	case Path + "token":
		if r.Method != http.MethodPost {
			return writeError(w, http.StatusMethodNotAllowed, "invalid_request",
				"mock-oauth: token endpoint requires POST")
		}
		return a.Token(w, r)
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		return json.NewEncoder(w).Encode(map[string]string{
			"error":             "unknown_endpoint",
			"error_description": fmt.Sprintf("No mock-oauth endpoint at %s", r.URL.Path),
		})
	}
}

// AuthorizeForm handles GET /mock-oauth/authorize.
// Renders a consent screen showing the requesting app and scopes.
func (a *Handler) AuthorizeForm(w http.ResponseWriter, r *http.Request) error {
	responseType := r.FormValue("response_type")
	if responseType == "" {
		return writeError(w, http.StatusBadRequest, "invalid_request", errMissingResponseType.Error())
	}
	if responseType != "code" {
		return writeError(w, http.StatusBadRequest, "unsupported_response_type", errInvalidResponseType.Error())
	}

	clientID := r.FormValue("client_id")
	if clientID == "" {
		return writeError(w, http.StatusBadRequest, "invalid_request", errMissingClientID.Error())
	}
	if strings.Contains(clientID, "|") {
		return writeError(w, http.StatusBadRequest, "invalid_request", errInvalidClientIDChar.Error())
	}

	redirectURI := r.FormValue("redirect_uri")
	if redirectURI == "" {
		return writeError(w, http.StatusBadRequest, "invalid_request", errMissingRedirectURI.Error())
	}
	if strings.Contains(redirectURI, "#") {
		return writeError(w, http.StatusBadRequest, "invalid_request", errRedirectURIFragment.Error())
	}

	scope := r.FormValue("scope")
	if strings.Contains(scope, "|") {
		return writeError(w, http.StatusBadRequest, "invalid_request", errInvalidScopeChar.Error())
	}
	var scopeItems h.Frag
	if scope != "" {
		for _, s := range strings.Split(scope, ",") {
			scopeItems = append(scopeItems, &h.Li{Class: "scope-item", Inner: h.String(s)})
		}
	}

	var scopeSection h.HTML
	if len(scopeItems) > 0 {
		scopeSection = &h.Ul{Class: "scope-list", Inner: scopeItems}
	} else {
		scopeSection = &h.Div{Class: "no-scopes", Inner: h.String("No specific scopes requested")}
	}

	page := &h.Document{
		Lang: "en",
		Inner: h.Frag{
			&h.Head{Inner: h.Frag{
				&h.Meta{Charset: "utf-8"},
				&h.Node{Tag: "meta", Attributes: h.Attributes{
					"name": "viewport", "content": "width=device-width, initial-scale=1",
				}, SelfClosing: true},
				&h.Node{Tag: "title", Inner: h.String("Authorize — Mock OAuth Provider")},
				&h.Link{Rel: "stylesheet", Type: "text/css", HREF: "/public/css/mockoauth.css"},
			}},
			&h.Body{Inner: &h.Div{Class: "card", Inner: h.Frag{
				&h.Div{Class: "header", Inner: h.Frag{
					&h.H1{Inner: h.String("Authorize " + clientID + " for FBRell")},
					&h.Div{Class: "subtitle", Inner: h.String("Mock OAuth Provider")},
				}},
				&h.Div{Class: "body", Inner: h.Frag{
					&h.Div{Class: "app-info", Inner: h.Frag{
						&h.Div{Class: "app-icon", Inner: h.Unsafe("&#128273;")},
						&h.Div{Class: "app-id", Inner: h.String("Client ID: " + clientID)},
					}},
					&h.Div{Class: "section-label",
						Inner: h.String("This app is requesting access to:")},
					scopeSection,
					&h.Form{Method: "post", Action: "/mock-oauth/authorize", Inner: h.Frag{
						hiddenInput("client_id", clientID),
						hiddenInput("redirect_uri", redirectURI),
						hiddenInput("state", r.FormValue("state")),
						hiddenInput("scope", scope),
						hiddenInput("behavior", r.FormValue("behavior")),
						&h.Div{Class: "actions", Inner: h.Frag{
							&h.Node{Tag: "button", Attributes: h.Attributes{
								"type": "button", "class": "btn btn-deny",
								"onclick": "window.close()",
							}, Inner: h.String("Cancel")},
							&h.Node{Tag: "button", Attributes: h.Attributes{
								"type": "submit", "name": "action", "value": "authorize",
								"class": "btn btn-authorize",
							}, Inner: h.String("Confirm")},
						}},
					}},
				}},
				&h.Div{Class: "mock-badge",
					Inner: h.String("⚠ This is a mock OAuth provider for testing only. " +
						"No real authentication occurs.")},
			}}},
		},
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err := h.Write(r.Context(), w, page)
	return err
}

// AuthorizeSubmit handles POST /mock-oauth/authorize.
// Processes the user's consent decision and redirects with a code.
func (a *Handler) AuthorizeSubmit(w http.ResponseWriter, r *http.Request) error {
	action := r.FormValue("action")
	if action != "authorize" {
		return writeError(w, http.StatusBadRequest, "invalid_request",
			"mock-oauth: action must be 'authorize'")
	}

	clientID := r.FormValue("client_id")
	if clientID == "" {
		return writeError(w, http.StatusBadRequest, "invalid_request", errMissingClientID.Error())
	}
	if strings.Contains(clientID, "|") {
		return writeError(w, http.StatusBadRequest, "invalid_request", errInvalidClientIDChar.Error())
	}

	redirectURI := r.FormValue("redirect_uri")
	if redirectURI == "" {
		return writeError(w, http.StatusBadRequest, "invalid_request", errMissingRedirectURI.Error())
	}
	if strings.Contains(redirectURI, "#") {
		return writeError(w, http.StatusBadRequest, "invalid_request", errRedirectURIFragment.Error())
	}

	state := r.FormValue("state")
	scope := r.FormValue("scope")
	if strings.Contains(scope, "|") {
		return writeError(w, http.StatusBadRequest, "invalid_request", errInvalidScopeChar.Error())
	}
	behavior := parseBehavior(r.FormValue("behavior"))

	code := BuildCode(clientID, scope, behavior)

	// RFC 6749 §3.1.2 allows redirect URIs to carry existing query params.
	// Parse the URI and merge our params into its query string so we append
	// with "&" rather than introducing a second "?".
	u, err := url.Parse(redirectURI)
	if err != nil {
		return writeError(w, http.StatusBadRequest, "invalid_request", "mock-oauth: invalid redirect_uri")
	}
	q := u.Query()
	q.Set("code", code)
	if state != "" {
		q.Set("state", state)
	}
	u.RawQuery = q.Encode()
	http.Redirect(w, r, u.String(), http.StatusFound)
	return nil
}

// Token handles POST /mock-oauth/token.
// Authenticates the client per RFC 6749 §2.3.1, then exchanges the mock
// authorization code for a mock access token embedding the granted scopes.
func (h *Handler) Token(w http.ResponseWriter, r *http.Request) error {
	grantType := r.FormValue("grant_type")
	if grantType != "" && grantType != "authorization_code" {
		return writeError(w, http.StatusBadRequest, "unsupported_grant_type", errInvalidGrantType.Error())
	}

	clientID, clientSecret, err := extractClientCredentials(r)
	if err != nil {
		// Auth failures (missing creds) get 401 invalid_client; malformed
		// requests (Basic + form-body conflict) get 400 invalid_request. We
		// don't distinguish "Basic was attempted" per RFC 6749 §5.2 because
		// we don't send WWW-Authenticate, so the spec's retry mechanism
		// doesn't apply here.
		switch {
		case errors.Is(err, errMissingClientID), errors.Is(err, errMissingClientSecret):
			return writeError(w, http.StatusUnauthorized, "invalid_client", err.Error())
		default:
			return writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		}
	}
	if !ValidateClient(clientID, clientSecret) {
		return writeError(w, http.StatusUnauthorized, "invalid_client", errInvalidClientCredentials.Error())
	}

	code := r.FormValue("code")
	if code == "" {
		return writeError(w, http.StatusBadRequest, "invalid_request", errMissingCode.Error())
	}

	codeClientID, scope, behavior, err := ParseCode(code)
	if err != nil {
		return writeError(w, http.StatusBadRequest, "invalid_grant", errInvalidCode.Error())
	}

	// Per RFC 6749 §4.1.3: ensure the authorization code was issued to
	// the authenticated client.
	if codeClientID != clientID {
		return writeError(w, http.StatusBadRequest, "invalid_grant", errClientIDMismatch.Error())
	}

	switch behavior {
	case BehaviorExpired:
		return writeError(w, http.StatusBadRequest, "invalid_grant", errExpiredCode.Error())
	case BehaviorInvalid:
		return writeError(w, http.StatusUnauthorized, "invalid_client", "mock-oauth: invalid client credentials")
	}

	token := buildToken(clientID, scope)

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(tokenResponse{
		AccessToken: token,
		TokenType:   "bearer",
		ExpiresIn:   3600,
		Scope:       scope,
	})
}

// extractClientCredentials reads OAuth 2.0 client credentials per RFC 6749
// §2.3.1. Prefers HTTP Basic Auth; falls back to form-body parameters.
// If both are present, they MUST match — otherwise this returns an error.
// Returns errMissingClientID when no credentials are supplied.
func extractClientCredentials(r *http.Request) (clientID, clientSecret string, err error) {
	basicID, basicSecret, hasBasic := r.BasicAuth()
	formID := r.PostFormValue("client_id")
	formSecret := r.PostFormValue("client_secret")

	switch {
	case hasBasic && (formID != "" || formSecret != ""):
		// Per RFC 6749 §2.3.1 clients SHOULD NOT use multiple methods. We
		// allow it only when the values match, to ease testing.
		if formID != "" && formID != basicID {
			return "", "", errCredentialsConflict
		}
		if formSecret != "" && formSecret != basicSecret {
			return "", "", errCredentialsConflict
		}
		return basicID, basicSecret, nil
	case hasBasic:
		return basicID, basicSecret, nil
	case formID != "" && formSecret != "":
		return formID, formSecret, nil
	case formID != "":
		return "", "", errMissingClientSecret
	default:
		return "", "", errMissingClientID
	}
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope,omitempty"`
}

// buildToken creates a deterministic, human-readable access token.
// Format: mock_token|{clientID}|{scope}
func buildToken(clientID, scope string) string {
	parts := []string{"mock_token", clientID}
	if scope != "" {
		parts = append(parts, scope)
	} else {
		parts = append(parts, "noscope")
	}
	return strings.Join(parts, "|")
}

// BuildCode creates a deterministic, human-readable authorization code.
// Format: mock_code|{clientID}|{scope}|{behavior}|{timestamp}
func BuildCode(clientID, scope string, behavior TokenBehavior) string {
	parts := []string{"mock_code", clientID}
	if scope != "" {
		parts = append(parts, scope)
	} else {
		parts = append(parts, "noscope")
	}
	parts = append(parts, string(behavior))
	parts = append(parts, fmt.Sprintf("%d", time.Now().Unix()))
	return strings.Join(parts, "|")
}

// ParseCode extracts client_id, scope, and behavior from a mock authorization code.
func ParseCode(code string) (clientID, scope string, behavior TokenBehavior, err error) {
	if !strings.HasPrefix(code, "mock_code|") {
		return "", "", "", errInvalidCode
	}

	trimmed := strings.TrimPrefix(code, "mock_code|")
	parts := strings.Split(trimmed, "|")
	if len(parts) < 3 {
		return "", "", "", errInvalidCode
	}

	clientID = parts[0]
	scope = parts[1]
	if scope == "noscope" {
		scope = ""
	}
	behavior = TokenBehavior(parts[2])

	return clientID, scope, behavior, nil
}

func parseBehavior(s string) TokenBehavior {
	switch TokenBehavior(s) {
	case BehaviorExpired:
		return BehaviorExpired
	case BehaviorInvalid:
		return BehaviorInvalid
	default:
		return BehaviorValid
	}
}

func hiddenInput(name, value string) h.HTML {
	return &h.Input{Type: "hidden", Name: name, Value: value}
}

func writeError(w http.ResponseWriter, status int, errorCode, description string) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(map[string]string{
		"error":             errorCode,
		"error_description": description,
	})
}
