// Package viewcontext provides context related HTTP Handlers for Rell.
package viewcontext

import (
	"encoding/json"
	"github.com/nshah/rell/context"
	"github.com/nshah/rell/view"
	"log"
	"net/http"
)

// Handler for /info/ to see a JSON view of some server context.
func Info(w http.ResponseWriter, r *http.Request) {
	context, err := context.FromRequest(r)
	if err != nil {
		view.Error(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	info := map[string]interface{}{
		"request": map[string]interface{}{
			"method": r.Method,
			"form":   r.Form,
			"url": map[string]interface{}{
				"path":  r.URL.Path,
				"query": r.URL.RawQuery,
			},
		},
		"context":    context,
		"pageTabURL": context.PageTabURL("/"),
		"canvasURL":  context.CanvasURL("/"),
		"sdkURL":     context.SdkURL(),
		"version":    "3.0.9",
	}
	out, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		log.Printf("Error json.MarshalIndent for /info/: %s", err)
	}
	w.Write(out)
	w.Write([]byte("\n"))
}
