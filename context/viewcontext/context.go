// Package viewcontext provides context related HTTP Handlers for Rell.
package viewcontext

import (
	"encoding/json"
	"github.com/daaku/go.h"
	"github.com/daaku/rell/context"
	"github.com/daaku/rell/view"
	"log"
	"net/http"
	"strings"
)

var version string

// Prints HTMLized JSON for Browsers & plain text for others.
func humanJSON(v interface{}, w http.ResponseWriter, r *http.Request) {
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Printf("Error json.MarshalIndent: %s", err)
	}
	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		view.Write(w, r, &h.Document{
			Inner: &h.Frag{
				&h.Head{
					Inner: &h.Frag{
						&h.Meta{Charset: "utf-8"},
						&h.Title{h.String("Dump")},
					},
				},
				&h.Body{
					Inner: &h.Pre{
						Inner: h.String(string(out)),
					},
				},
			},
		})
	} else {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write(out)
		w.Write([]byte("\n"))
	}
}

func headerMap(h http.Header) map[string]string {
	r := make(map[string]string)
	for name, value := range h {
		r[name] = value[0]
	}
	return r
}

// Handler for /info/ to see a JSON view of some server context.
func Info(w http.ResponseWriter, r *http.Request) {
	context, err := context.FromRequest(r)
	if err != nil {
		view.Error(w, r, err)
		return
	}
	info := map[string]interface{}{
		"request": map[string]interface{}{
			"method": r.Method,
			"form":   r.Form,
			"url": map[string]interface{}{
				"path":  r.URL.Path,
				"query": r.URL.RawQuery,
			},
			"headers": headerMap(r.Header),
		},
		"context":    context,
		"pageTabURL": context.PageTabURL("/"),
		"canvasURL":  context.CanvasURL("/"),
		"sdkURL":     context.SdkURL(),
	}
	if version != "" {
		info["version"] = version
	}
	humanJSON(info, w, r)
}
