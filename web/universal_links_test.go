package web

import (
	"bytes"
	"encoding/json"
	"log"
	"mime"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

type unreadCallbackBody struct{}

func (unreadCallbackBody) Read([]byte) (int, error) {
	panic("authorization request body must not be read")
}

func TestUniversalLinkCallbackIsCredentialIndependent(t *testing.T) {
	for _, method := range []string{
		http.MethodGet, http.MethodHead, http.MethodPost,
		http.MethodPut, http.MethodDelete, http.MethodOptions,
	} {
		t.Run(method, func(t *testing.T) {
			var logs bytes.Buffer
			h := &Handler{Logger: log.New(&logs, "", 0)}
			r := httptest.NewRequest(method,
				"https://www.fbrell.com"+universalLinkCallbackPath+
					"?code=fake_code&state=fake_state&access_token=fake_token"+
					"&signed_request=invalid&error_description=fake_error", unreadCallbackBody{})
			r.Header.Set("Cookie", "signed_request=invalid")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			switch method {
			case http.MethodGet:
				if w.Code != http.StatusOK || w.Body.String() != universalLinkCallbackNotice {
					t.Fatalf("got %d and %q, want 200 and the fixed iOS notice", w.Code, w.Body.String())
				}
			case http.MethodHead:
				if w.Code != http.StatusOK || w.Body.Len() != 0 {
					t.Fatalf("got %d and %q, want 200 without a body", w.Code, w.Body.String())
				}
			default:
				if w.Code != http.StatusMethodNotAllowed || w.Header().Get("Allow") != "GET, HEAD" {
					t.Fatalf("got %d and Allow %q, want 405 and GET, HEAD", w.Code, w.Header().Get("Allow"))
				}
				if bytes.Contains(w.Body.Bytes(), []byte("fake_")) {
					t.Fatal("callback echoed authorization data")
				}
			}
			if logs.Len() != 0 || w.Header().Get("Location") != "" {
				t.Fatal("callback must not enter error logging or redirect")
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

func TestAppleAssociationContentTypeAndCallbackContract(t *testing.T) {
	h := staticFileHandler(t, http.Dir("../public"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, appleAssociationPath, nil))
	contentType, _, err := mime.ParseMediaType(w.Header().Get("Content-Type"))
	if err != nil || contentType != "application/json" {
		t.Fatalf("got Content-Type %q, want application/json", w.Header().Get("Content-Type"))
	}
	if w.Header().Get("Location") != "" || w.Header().Get("WWW-Authenticate") != "" {
		t.Fatal("Apple association must be public without redirects")
	}
	var association struct {
		AppLinks struct {
			Details []struct {
				AppIDs     []string `json:"appIDs"`
				Components []struct {
					Path    string `json:"/"`
					Exclude bool   `json:"exclude"`
				} `json:"components"`
			} `json:"details"`
		} `json:"applinks"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &association); err != nil {
		t.Fatal(err)
	}
	details := association.AppLinks.Details
	if len(details) != 1 || len(details[0].AppIDs) != 1 ||
		details[0].AppIDs[0] != "T84QZS65DQ.com.facebook.internal.sdk.samples.Hackbook.localDevelopment" {
		t.Fatal("unexpected Apple application identifier")
	}
	paths := make(map[string]bool)
	for _, component := range details[0].Components {
		if component.Exclude {
			t.Fatal("expected allowed callback components")
		}
		paths[component.Path] = true
	}
	if len(details[0].Components) != 2 || !paths[universalLinkCallbackPath] || !paths["/oauth/response/*"] {
		t.Fatalf("want exact iOS callback and preserved legacy path, got %v", paths)
	}
}

func TestAppleAssociationBypassesApplicationMiddleware(t *testing.T) {
	const body = "{\"applinks\":{\"details\":[]}}\n"
	for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodPost} {
		t.Run(method, func(t *testing.T) {
			h := &Handler{PublicFS: http.FS(fstest.MapFS{
				".well-known/apple-app-site-association": &fstest.MapFile{Data: []byte(body)},
			})}
			r := httptest.NewRequest(method, appleAssociationPath+"?signed_request=invalid", unreadCallbackBody{})
			r.Header.Set("Cookie", "signed_request=invalid")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if method == http.MethodPost {
				if w.Code != http.StatusMethodNotAllowed || w.Header().Get("Allow") != "GET, HEAD" {
					t.Fatalf("got %d and Allow %q", w.Code, w.Header().Get("Allow"))
				}
				return
			}
			if w.Code != http.StatusOK || w.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("got %d and Content-Type %q", w.Code, w.Header().Get("Content-Type"))
			}
			if method == http.MethodGet && w.Body.String() != body {
				t.Fatalf("got %q, want configured file %q", w.Body.String(), body)
			}
			if method == http.MethodHead && w.Body.Len() != 0 {
				t.Fatal("HEAD must not return a body")
			}
		})
	}
}

func TestMissingAppleAssociationReturnsNotFound(t *testing.T) {
	h := &Handler{PublicFS: http.FS(fstest.MapFS{})}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, appleAssociationPath, nil))
	if w.Code != http.StatusNotFound || w.Header().Get("Location") != "" {
		t.Fatalf("got %d and Location %q, want 404 without redirect", w.Code, w.Header().Get("Location"))
	}
}
