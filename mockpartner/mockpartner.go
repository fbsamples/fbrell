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

// Package mockpartner provides shared infrastructure for mock partner API
// endpoints. It handles Bearer token validation against the mock OAuth
// provider's token format (mock_token|{clientID}|{scope}).
package mockpartner

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

var (
	ErrMissingAuth  = errors.New("mockpartner: missing Authorization header")
	ErrInvalidAuth  = errors.New("mockpartner: invalid Bearer token")
	ErrInvalidToken = errors.New("mockpartner: token is not a valid mock_token")
)

// TokenInfo holds the parsed contents of a mock OAuth access token.
type TokenInfo struct {
	ClientID string
	Scopes   []string
}

// ParseBearerToken extracts and validates a mock_token from the Authorization header.
// Expected format: "Bearer mock_token|{clientID}|{scope}" where scope is comma-separated.
func ParseBearerToken(r *http.Request) (*TokenInfo, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return nil, ErrMissingAuth
	}

	if !strings.HasPrefix(auth, "Bearer ") {
		return nil, ErrInvalidAuth
	}
	token := strings.TrimPrefix(auth, "Bearer ")

	if !strings.HasPrefix(token, "mock_token|") {
		return nil, ErrInvalidToken
	}
	trimmed := strings.TrimPrefix(token, "mock_token|")
	parts := strings.SplitN(trimmed, "|", 2)
	if len(parts) < 2 {
		return nil, ErrInvalidToken
	}

	clientID := parts[0]
	scopeStr := parts[1]
	var scopes []string
	if scopeStr != "" && scopeStr != "noscope" {
		scopes = strings.Split(scopeStr, ",")
	}

	return &TokenInfo{ClientID: clientID, Scopes: scopes}, nil
}

// WriteError writes a JSON error response matching the OAuth error format.
func WriteError(w http.ResponseWriter, status int, errorCode, description string) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(map[string]string{
		"error":             errorCode,
		"error_description": description,
	})
}
