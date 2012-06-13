// Package stats implements Rell insights.
package stats

import (
	"flag"
	"github.com/stathat/stathatgo"
	"log"
	"net/http"
	"time"
)

var (
	ezkey   = flag.String("rell.stats.key", "", "The stathat ezkey.")
	verbose = flag.Bool(
		"rell.stats.verbose", false, "Enable verbose logging of stats.")
)

type Handler struct {
	Handler http.Handler
	Name    string
}

func Count(name string, count int) {
	if *verbose {
		log.Printf("stats.Count(%s, %d)", name, count)
	}
	err := stathat.PostEZCount(name, *ezkey, count)
	if err != nil {
		log.Printf("Failed to PostEZCount: %s", err)
	}
}

func Inc(name string) {
	Count(name, 1)
}

func Record(name string, value float64) {
	if *verbose {
		log.Printf("stats.Value(%s, %f)", name, value)
	}
	err := stathat.PostEZValue(name, *ezkey, value)
	if err != nil {
		log.Printf("Failed to PostEZCount: %s", err)
	}
}

func NewHandler(name string, handler http.Handler) *Handler {
	return &Handler{
		Handler: handler,
		Name:    name,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	Inc("web request")
	Inc("web request - method=" + r.Method)
	start := time.Now()
	h.Handler.ServeHTTP(w, r)
	Record("web request gen time", time.Since(start).Seconds())
}
