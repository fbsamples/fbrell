// Package web provides the main HTTP entrypoint for rell.
package web

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"path/filepath"
	"sync"

	"github.com/daaku/go.httpdev"
	"github.com/daaku/go.httpgzip"
	"github.com/daaku/go.signedrequest/appdata"
	"github.com/daaku/go.static"
	"github.com/daaku/go.viewvar"
	"github.com/facebookgo/fbapp"

	"github.com/daaku/rell/context/viewcontext"
	"github.com/daaku/rell/examples/viewexamples"
	"github.com/daaku/rell/oauth"
	"github.com/daaku/rell/og/viewog"
)

// The rell web application.
type App struct {
	ContextHandler  *viewcontext.Handler
	ExamplesHandler *viewexamples.Handler
	OgHandler       *viewog.Handler
	OauthHandler    *oauth.Handler
	Static          *static.Handler
	App             fbapp.App

	adminHandler     http.Handler
	adminHandlerOnce sync.Once
	mainHandler      http.Handler
	mainHandlerOnce  sync.Once
}

// Serve HTTP requests for the admin port.
func (a *App) AdminHandler(w http.ResponseWriter, r *http.Request) {
	a.adminHandlerOnce.Do(func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/vars/", viewvar.Json)
		mux.HandleFunc("/env/", a.envHandler)
		a.adminHandler = mux
	})
	a.adminHandler.ServeHTTP(w, r)
}

// Serve HTTP requests for the main port.
func (a *App) MainHandler(w http.ResponseWriter, r *http.Request) {
	a.mainHandlerOnce.Do(func() {
		const public = "/public/"

		mux := http.NewServeMux()
		mux.Handle(a.Static.HttpPath, a.Static)

		a.staticFile(mux, "/favicon.ico")
		a.staticFile(mux, "/f8.jpg")
		a.staticFile(mux, "/robots.txt")
		mux.Handle(public,
			http.StripPrefix(
				public, http.FileServer(http.Dir(a.Static.DiskPath))))

		mux.HandleFunc("/not_a_real_webpage", http.NotFound)
		mux.Handle("/info/", a.ContextHandler)
		mux.HandleFunc("/examples/", a.ExamplesHandler.List)
		mux.HandleFunc("/saved/", a.ExamplesHandler.Saved)
		mux.HandleFunc("/raw/", a.ExamplesHandler.Raw)
		mux.HandleFunc("/simple/", a.ExamplesHandler.Simple)
		mux.HandleFunc("/", a.ExamplesHandler.Example)
		mux.HandleFunc("/og/", a.OgHandler.Values)
		mux.HandleFunc("/rog/", a.OgHandler.Base64)
		mux.HandleFunc("/rog-redirect/", a.OgHandler.Redirect)
		mux.Handle(oauth.Path, a.OauthHandler)
		mux.HandleFunc("/sleep/", httpdev.Sleep)

		var handler http.Handler
		handler = &appdata.Handler{
			Handler: mux,
			Secret:  a.App.SecretByte(),
		}
		handler = httpgzip.NewHandler(handler)
		a.mainHandler = handler
	})
	a.mainHandler.ServeHTTP(w, r)
}

// binds a path to a single file
func (a *App) staticFile(mux *http.ServeMux, name string) {
	abs := filepath.Join(a.Static.DiskPath, name)
	mux.HandleFunc(name, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, abs)
	})
}

func (a *App) envHandler(w http.ResponseWriter, r *http.Request) {
	for _, s := range os.Environ() {
		fmt.Fprintln(w, s)
	}
}
