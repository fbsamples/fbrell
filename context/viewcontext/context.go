// Package viewcontext provides context related HTTP Handlers for Rell.
package viewcontext

import (
	"encoding/json"
	"github.com/nshah/rell/context"
	"github.com/nshah/rell/view"
	"net/http"
)

// Handler for /info/ to see a JSON view of some server context.
func Info(w http.ResponseWriter, r *http.Request) {
	context, err := context.FromRequest(r)
	if err != nil {
		view.Error(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	info := map[string]interface{}{
		"context":    context,
		"pageTabURL": context.PageTabURL(),
		"canvasURL":  context.CanvasURL(),
		"channelURL": context.ChannelURL(),
		"sdkURL":     context.SdkURL(),
		"version":    "3.0.6",
	}
	json.NewEncoder(w).Encode(info)
}
