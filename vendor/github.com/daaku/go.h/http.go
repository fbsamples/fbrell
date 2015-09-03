// +build !js

package h

import (
	"log"
	"net/http"

	"golang.org/x/net/context"
)

// Writes a HTML response and writes errors on failure.
func WriteResponse(ctx context.Context, w http.ResponseWriter, r *http.Request, html HTML) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r.Method != "HEAD" {
		_, err := Write(ctx, w, html)
		if err != nil {
			log.Printf("Error writing HTML for URL: %s: %s", r.URL, err)
			Write(ctx, w, String("FATAL ERROR"))
		}
	}
}
