package web

import (
	"io"
	"net/http"
)

const (
	oauthCallbackPath         = "/app-links/callback"
	universalLinkCallbackPath = "/universal-links/callback"
	appleAssociationPath      = "/.well-known/apple-app-site-association"
)

const oauthCallbackNotice = "This HTTPS callback reached the web server instead of the Android app.\n" +
	"No authorization data is displayed or processed here.\n"

const universalLinkCallbackNotice = "This HTTPS callback reached the web server instead of the iOS app.\n" +
	"No authorization data is displayed or processed here.\n"

func serveOAuthCallback(w http.ResponseWriter, r *http.Request, notice string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")

	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.Method == http.MethodHead {
		w.WriteHeader(http.StatusOK)
		return
	}
	_, _ = io.WriteString(w, notice)
}
