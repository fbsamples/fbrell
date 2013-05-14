// Package viewcontext provides context related HTTP Handlers for Rell.
package viewcontext

import (
	"net/http"

	"github.com/daaku/go.httpdev"

	"github.com/daaku/rell/context"
	"github.com/daaku/rell/view"
)

var version string

// Handler for /info/ to see a JSON view of some server context.
func Info(w http.ResponseWriter, r *http.Request) {
	context, err := context.FromRequest(r)
	if err != nil {
		view.Error(w, r, err)
		return
	}
	info := map[string]interface{}{
		"context":    context,
		"pageTabURL": context.PageTabURL("/"),
		"canvasURL":  context.CanvasURL("/"),
		"sdkURL":     context.SdkURL(),
		"version":    version,
	}
	httpdev.Info(info, w, r)
}
