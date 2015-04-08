package adminweb

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"path"
	"sync"

	"github.com/daaku/rell/internal/github.com/daaku/go.httpdev"
	"github.com/daaku/rell/internal/github.com/daaku/go.trustforward"
	"github.com/daaku/rell/internal/github.com/daaku/go.viewvar"
)

type Handler struct {
	Forwarded *trustforward.Forwarded
	SkipHTTPS bool
	Path      string

	mux  http.Handler
	once sync.Once
}

var httpsRequired = []byte("https required\n")

// Serve HTTP requests for the admin port.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.SkipHTTPS && h.Forwarded.Scheme(r) != "https" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(httpsRequired))
		return
	}
	h.once.Do(func() {
		root := path.Join("/", h.Path) + "/"
		mux := http.NewServeMux()
		mux.HandleFunc(root+"debug/pprof/", pprof.Index)
		mux.HandleFunc(root+"debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc(root+"debug/pprof/profile", pprof.Profile)
		mux.HandleFunc(root+"debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc(root+"vars/", viewvar.Json)
		mux.HandleFunc(root+"env/", h.envHandler)
		mux.HandleFunc(root+"sleep/", httpdev.Sleep)
		h.mux = mux
	})
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) envHandler(w http.ResponseWriter, r *http.Request) {
	for _, s := range os.Environ() {
		fmt.Fprintln(w, s)
	}
}
