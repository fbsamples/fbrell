// +build !js

package h

import (
	"log"
	"net/http"
)

// Writes a HTML response and writes errors on failure.
func WriteResponse(w http.ResponseWriter, r *http.Request, html HTML) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r.Method != "HEAD" {
		_, err := Write(w, html)
		if err != nil {
			log.Printf("Error writing HTML for URL: %s: %s", r.URL, err)
			Write(w, String("FATAL ERROR"))
		}
	}
}
