// Package viewcontext provides context related HTTP Handlers for Rell.
package viewcontext

import (
	"net/http"

	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.httpdev"
	"github.com/daaku/rell/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/daaku/rell/rellenv"
)

var version string

type Handler struct {
	ContextParser *rellenv.Parser
}

// Handler for /info/ to see a JSON view of some server context.
func (h *Handler) Info(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	context, err := h.ContextParser.FromRequest(r)
	if err != nil {
		return err
	}
	info := map[string]interface{}{
		"context":    context,
		"pageTabURL": context.PageTabURL("/"),
		"canvasURL":  context.CanvasURL("/"),
		"sdkURL":     context.SdkURL(),
		"version":    version,
	}
	httpdev.Info(info, w, r)
	return nil
}
