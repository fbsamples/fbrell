package web

import (
	"encoding/json"
	"io/fs"
	"mime"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"
	"testing/fstest"

	static "github.com/daaku/go.static"
	trustforward "github.com/daaku/go.trustforward"
	"github.com/facebookgo/fbapp"
	"github.com/fbsamples/fbrell/adminweb"
	"github.com/fbsamples/fbrell/rellenv"
)

const assetLinksPath = "/.well-known/assetlinks.json"

type staticAppNamespace struct{}

func (staticAppNamespace) Get(uint64) string { return "" }

func staticFileHandler(t *testing.T, publicFS http.FileSystem) *Handler {
	t.Helper()
	app := fbapp.New(1, "", "")
	h := &Handler{
		App:          app,
		PublicFS:     publicFS,
		Static:       &static.Handler{Path: "/static/"},
		AdminHandler: &adminweb.Handler{},
		EnvParser: &rellenv.Parser{
			App:          app,
			AppNSFetcher: staticAppNamespace{},
			Forwarded:    &trustforward.Forwarded{},
		},
	}
	if err := h.Init(); err != nil {
		t.Fatal(err)
	}
	return h
}

func TestAssetLinksEndpoint(t *testing.T) {
	h := staticFileHandler(t, http.Dir("../public"))
	w := httptest.NewRecorder()
	// No cookies or authorization headers: Android must fetch this publicly.
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet,
		"https://www.fbrell.com"+assetLinksPath, nil))

	if w.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusOK)
	}
	contentType, _, err := mime.ParseMediaType(w.Header().Get("Content-Type"))
	if err != nil || contentType != "application/json" {
		t.Fatalf("got Content-Type %q, want application/json", w.Header().Get("Content-Type"))
	}
	if location := w.Header().Get("Location"); location != "" {
		t.Fatalf("association endpoint must not redirect, got %q", location)
	}
	if challenge := w.Header().Get("WWW-Authenticate"); challenge != "" {
		t.Fatalf("association endpoint must not require authentication, got %q", challenge)
	}
	expected, err := os.ReadFile("../public/.well-known/assetlinks.json")
	if err != nil {
		t.Fatal(err)
	}
	if w.Body.String() != string(expected) {
		t.Fatalf("got body %q, want the association file %q", w.Body.String(), expected)
	}

	var statements []struct {
		Relation []string `json:"relation"`
		Target   struct {
			Namespace              string   `json:"namespace"`
			PackageName            string   `json:"package_name"`
			SHA256CertFingerprints []string `json:"sha256_cert_fingerprints"`
		} `json:"target"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &statements); err != nil {
		t.Fatal(err)
	}
	if statements == nil {
		t.Fatal("association file must be a JSON array, not null")
	}
	// An empty array is intentionally non-authorizing until the APK is ready.
	// When populated, reject incomplete statements and placeholder fingerprints.
	fingerprint := regexp.MustCompile(`(?i)^(?:[0-9a-f]{2}:){31}[0-9a-f]{2}$`)
	for _, statement := range statements {
		if len(statement.Relation) != 1 || statement.Relation[0] != "delegate_permission/common.handle_all_urls" {
			t.Fatalf("unexpected relation: %v", statement.Relation)
		}
		if statement.Target.Namespace != "android_app" || statement.Target.PackageName != "com.facebook.hackbook.internal.current" {
			t.Fatalf("unexpected Android target: %+v", statement.Target)
		}
		if len(statement.Target.SHA256CertFingerprints) == 0 {
			t.Fatal("a published association must include the installed APK's signing fingerprint")
		}
		for _, value := range statement.Target.SHA256CertFingerprints {
			if !fingerprint.MatchString(value) {
				t.Fatalf("invalid signing certificate SHA-256 fingerprint %q", value)
			}
		}
	}
}

func TestAssetLinksUsesConfiguredPublicDirectory(t *testing.T) {
	const body = "[]\n"
	h := staticFileHandler(t, http.FS(fstest.MapFS{
		".well-known/assetlinks.json": &fstest.MapFile{Data: []byte(body)},
	}))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, assetLinksPath, nil))
	if w.Code != http.StatusOK || w.Body.String() != body {
		t.Fatalf("got status %d and body %q, want 200 and %q", w.Code, w.Body.String(), body)
	}
}

func TestMissingAssetLinksReturnsNotFound(t *testing.T) {
	h := staticFileHandler(t, http.FS(fstest.MapFS{}))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, assetLinksPath, nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusNotFound)
	}
	if location := w.Header().Get("Location"); location != "" {
		t.Fatalf("missing association file must not redirect, got %q", location)
	}
}

func TestAppleAssociationEndpointUnchanged(t *testing.T) {
	h := staticFileHandler(t, http.Dir("../public"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet,
		"/.well-known/apple-app-site-association", nil))
	expected, err := fs.ReadFile(os.DirFS("../public"), ".well-known/apple-app-site-association")
	if err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusOK || w.Body.String() != string(expected) {
		t.Fatalf("got status %d and body %q, want 200 and the unchanged Apple association file", w.Code, w.Body.String())
	}
}
