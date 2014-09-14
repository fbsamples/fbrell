// Package httpstats provides a HTTP handler that will keep track of
// some useful request statistics.
package httpstats

import (
	"net/http"
	"time"
)

type Stats interface {
	Count(name string, count int)
	Record(name string, value float64)
}

type Handler struct {
	Handler http.Handler
	Name    string
	Stats   Stats
}

func NewHandler(name string, handler http.Handler) *Handler {
	return &Handler{
		Handler: handler,
		Name:    name,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Stats.Count("web request", 1)
	h.Stats.Count("web request - method="+r.Method, 1)
	start := time.Now()
	h.Handler.ServeHTTP(w, r)
	h.Stats.Record("web request gen time", float64(time.Since(start).Nanoseconds()))
}
