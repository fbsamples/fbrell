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
// GET/POST /mock-oauth/authorize renders a consent screen and issues an
// authorization code on user approval.
//
// Codes are non-cryptographic, human-readable strings encoding the client ID,
// granted scopes, and configurable behavior (valid/expired/invalid).
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

const Path = "/mock-oauth/"

var (
	errMissingClientID     = errors.New("mock-oauth: missing client_id parameter")
	errMissingRedirectURI  = errors.New("mock-oauth: missing redirect_uri parameter")
	errMissingResponseType = errors.New("mock-oauth: missing response_type parameter")
	errInvalidResponseType = errors.New("mock-oauth: response_type must be 'code'")
	errRedirectURIFragment = errors.New("mock-oauth: redirect_uri must not contain a fragment")
	errInvalidClientIDChar = errors.New("mock-oauth: client_id must not contain '|'")
	errInvalidScopeChar    = errors.New("mock-oauth: scope values must not contain '|'")
)

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
