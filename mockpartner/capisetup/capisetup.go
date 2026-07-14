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

// Package capisetup implements mock partner API endpoints for the CAPI Setup
// flow, used to validate clients that integrate with a partner's CAPI Setup API.
//
// Four endpoints (two flat, two path-parameterised):
//   - GET  /mock-partner/capi-setup/business_contexts — returns the set of
//     business contexts (e.g., available stores) the authenticated client
//     has access to.
//   - POST /mock-partner/capi-setup/connect_and_share_pixel — accepts a pixel
//     identifier scoped to a business context, returns a confirmation.
//   - GET  /mock-partner/capi-setup/{user_id}/business_contexts — same as the
//     flat variant, but captures a partner-side user identifier from the URL
//     path and echoes it in the response body. Enables testing of clients
//     that substitute a stored partner-routing id into outbound URLs.
//   - POST /mock-partner/capi-setup/{user_id}/connect_and_share_pixel — same
//     as the flat variant, but captures user_id from the URL path and echoes
//     it in the response body.
//
// All endpoints require a Bearer token issued by the mock OAuth provider.
package capisetup

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/fbsamples/fbrell/mockpartner"
)

const Path = "/mock-partner/capi-setup/"

// RequiredScope is the OAuth scope a Bearer token must carry to access any
// capisetup endpoint. The mock OAuth provider issues tokens with the scopes
// requested at the authorize step, so callers must request scope=write_capi_setup.
const RequiredScope = "write_capi_setup"

var (
	errMissingPixelID    = errors.New("capisetup: missing pixel_id parameter")
	errMissingBusinessID = errors.New("capisetup: missing business_id parameter")
	errUnknownContextID  = errors.New("capisetup: unknown context_id")
	errInsufficientScope = errors.New("capisetup: token missing required scope " + RequiredScope)
)

// mockBusinessContexts is the in-memory "database" of business contexts. It is
// the single source of truth for both business_contexts (which returns it) and
// connect_and_share_pixel (which validates context_id against it).
var mockBusinessContexts = []BusinessContext{
	{
		ContextID:      "123",
		ContextName:    "Test Store 1",
		ContextSubtext: "Primary test store",
	},
	{
		ContextID:   "456",
		ContextName: "Test Store 2",
	},
}

// isKnownContextID reports whether id matches a context in mockBusinessContexts.
func isKnownContextID(id string) bool {
	for _, c := range mockBusinessContexts {
		if c.ContextID == id {
			return true
		}
	}
	return false
}

// hasScope reports whether scopes contains the given scope.
func hasScope(scopes []string, scope string) bool {
	for _, s := range scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// Handler serves mock CAPI Setup partner API endpoints.
type Handler struct{}

// Handle routes requests to the appropriate CAPI Setup endpoint.
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) error {
	token, err := mockpartner.ParseBearerToken(r)
	if err != nil {
		return mockpartner.WriteError(w, http.StatusUnauthorized, "invalid_token", err.Error())
	}
	if !hasScope(token.Scopes, RequiredScope) {
		return mockpartner.WriteError(w, http.StatusForbidden, "insufficient_scope", errInsufficientScope.Error())
	}

	// Path shape: everything after Path is either "<endpoint>" (flat) or
	// "<user_id>/<endpoint>" (path-parameterised). Deeper nesting is 404.
	trimmed := strings.TrimPrefix(r.URL.Path, Path)
	segments := strings.Split(trimmed, "/")

	var userID, endpoint string
	switch len(segments) {
	case 1:
		endpoint = segments[0]
	case 2:
		userID, endpoint = segments[0], segments[1]
		if userID == "" {
			// e.g. "//business_contexts" — treat empty user_id as unknown
			return mockpartner.WriteError(w, http.StatusNotFound, "unknown_endpoint",
				fmt.Sprintf("No capi-setup endpoint at %s", r.URL.Path))
		}
	default:
		return mockpartner.WriteError(w, http.StatusNotFound, "unknown_endpoint",
			fmt.Sprintf("No capi-setup endpoint at %s", r.URL.Path))
	}

	switch endpoint {
	case "business_contexts":
		if r.Method != http.MethodGet {
			return mockpartner.WriteError(w, http.StatusMethodNotAllowed, "invalid_request",
				"capisetup: business_contexts requires GET")
		}
		if userID == "" {
			return h.BusinessContexts(w, r, token)
		}
		return h.BusinessContextsWithUserID(w, r, token, userID)
	case "connect_and_share_pixel":
		if r.Method != http.MethodPost {
			return mockpartner.WriteError(w, http.StatusMethodNotAllowed, "invalid_request",
				"capisetup: connect_and_share_pixel requires POST")
		}
		if userID == "" {
			return h.ConnectAndSharePixel(w, r, token)
		}
		return h.ConnectAndSharePixelWithUserID(w, r, token, userID)
	default:
		return mockpartner.WriteError(w, http.StatusNotFound, "unknown_endpoint",
			fmt.Sprintf("No capi-setup endpoint at %s", r.URL.Path))
	}
}

// BusinessContext represents a single business context (e.g., a store) that
// the authenticated client has access to.
type BusinessContext struct {
	ContextID      string `json:"context_id"`
	ContextName    string `json:"context_name"`
	ContextSubtext string `json:"context_subtext,omitempty"`
}

// BusinessContextsResponse is the wrapper object returned by GET business_contexts.
type BusinessContextsResponse struct {
	Contexts []BusinessContext `json:"contexts"`
}

// BusinessContexts handles GET /mock-partner/capi-setup/business_contexts.
//
// Required scope: write_capi_setup
//
// Request:
//
//	GET /mock-partner/capi-setup/business_contexts
//	Authorization: Bearer <token>
//
// Response (200):
//
//	{
//	  "contexts": [
//	    { "context_id": "123", "context_name": "Test Store 1", "context_subtext": "..." },
//	    { "context_id": "456", "context_name": "Test Store 2" }
//	  ]
//	}
func (h *Handler) BusinessContexts(w http.ResponseWriter, r *http.Request, _ *mockpartner.TokenInfo) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(BusinessContextsResponse{Contexts: mockBusinessContexts})
}

// ConnectAndSharePixelRequest holds the expected POST body.
type ConnectAndSharePixelRequest struct {
	ContextID  string `json:"context_id"`
	PixelID    string `json:"pixel_id"`
	BusinessID string `json:"business_id"`
	DebugID    string `json:"debug_id,omitempty"`
}

// ConnectAndSharePixelResponse is the mock success response.
type ConnectAndSharePixelResponse struct {
	Success bool `json:"success"`
}

// ConnectAndSharePixel handles POST /mock-partner/capi-setup/connect_and_share_pixel.
//
// Required scope: write_capi_setup
//
// Request:
//
//	POST /mock-partner/capi-setup/connect_and_share_pixel
//	Authorization: Bearer <token>
//	Content-Type: application/json
//
//	{
//	  "context_id":  "123",            // optional; if set, must match a known context
//	  "pixel_id":    "12345",          // required
//	  "business_id": "67890",          // required
//	  "debug_id":    "dbg_xyz"         // optional
//	}
//
// Response (200):
//
//	{ "success": true }
func (h *Handler) ConnectAndSharePixel(w http.ResponseWriter, r *http.Request, _ *mockpartner.TokenInfo) error {
	var req ConnectAndSharePixelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return mockpartner.WriteError(w, http.StatusBadRequest, "invalid_request",
			"capisetup: invalid JSON body")
	}

	if req.PixelID == "" {
		return mockpartner.WriteError(w, http.StatusBadRequest, "invalid_request", errMissingPixelID.Error())
	}
	if req.BusinessID == "" {
		return mockpartner.WriteError(w, http.StatusBadRequest, "invalid_request", errMissingBusinessID.Error())
	}
	if req.ContextID != "" && !isKnownContextID(req.ContextID) {
		return mockpartner.WriteError(w, http.StatusBadRequest, "invalid_request", errUnknownContextID.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(ConnectAndSharePixelResponse{Success: true})
}

// BusinessContextsWithUserIDResponse is the wrapper returned by
// GET business_contexts when a user_id path segment is present. It mirrors
// BusinessContextsResponse but adds the user_id echoed back from the URL so
// callers can verify the value round-tripped over the wire.
type BusinessContextsWithUserIDResponse struct {
	UserID   string            `json:"user_id"`
	Contexts []BusinessContext `json:"contexts"`
}

// BusinessContextsWithUserID handles
// GET /mock-partner/capi-setup/{user_id}/business_contexts.
//
// Required scope: write_capi_setup
//
// Request:
//
//	GET /mock-partner/capi-setup/1234/business_contexts
//	Authorization: Bearer <token>
//
// Response (200):
//
//	{
//	  "user_id": "1234",
//	  "contexts": [
//	    { "context_id": "123", "context_name": "Test Store 1", "context_subtext": "..." },
//	    { "context_id": "456", "context_name": "Test Store 2" }
//	  ]
//	}
func (h *Handler) BusinessContextsWithUserID(
	w http.ResponseWriter, r *http.Request, _ *mockpartner.TokenInfo, userID string,
) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(BusinessContextsWithUserIDResponse{
		UserID:   userID,
		Contexts: mockBusinessContexts,
	})
}

// ConnectAndSharePixelWithUserIDResponse is the wrapper returned by
// POST connect_and_share_pixel when a user_id path segment is present. It
// mirrors ConnectAndSharePixelResponse but adds the user_id echoed back from
// the URL so callers can verify the value round-tripped over the wire.
type ConnectAndSharePixelWithUserIDResponse struct {
	UserID  string `json:"user_id"`
	Success bool   `json:"success"`
}

// ConnectAndSharePixelWithUserID handles
// POST /mock-partner/capi-setup/{user_id}/connect_and_share_pixel.
//
// Required scope: write_capi_setup
//
// Same request body and validation as ConnectAndSharePixel; response adds
// the user_id captured from the URL.
//
// Response (200):
//
//	{ "user_id": "1234", "success": true }
func (h *Handler) ConnectAndSharePixelWithUserID(
	w http.ResponseWriter, r *http.Request, _ *mockpartner.TokenInfo, userID string,
) error {
	var req ConnectAndSharePixelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return mockpartner.WriteError(w, http.StatusBadRequest, "invalid_request",
			"capisetup: invalid JSON body")
	}

	if req.PixelID == "" {
		return mockpartner.WriteError(w, http.StatusBadRequest, "invalid_request", errMissingPixelID.Error())
	}
	if req.BusinessID == "" {
		return mockpartner.WriteError(w, http.StatusBadRequest, "invalid_request", errMissingBusinessID.Error())
	}
	if req.ContextID != "" && !isKnownContextID(req.ContextID) {
		return mockpartner.WriteError(w, http.StatusBadRequest, "invalid_request", errUnknownContextID.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(ConnectAndSharePixelWithUserIDResponse{
		UserID:  userID,
		Success: true,
	})
}
