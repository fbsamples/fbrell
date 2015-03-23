// Package web provides the main HTTP entrypoint for rell.
package web

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"sync"
	"time"

	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.httpdev"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.httpgzip"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.signedrequest/appdata"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.static"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.viewvar"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/facebookgo/fbapp"
	"github.com/daaku/rell/context/viewcontext"
	"github.com/daaku/rell/examples/viewexamples"
	"github.com/daaku/rell/oauth"
	"github.com/daaku/rell/og/viewog"
)

// The rell web application.
type App struct {
	ContextHandler      *viewcontext.Handler
	ExamplesHandler     *viewexamples.Handler
	OgHandler           *viewog.Handler
	OauthHandler        *oauth.Handler
	Static              *static.Handler
	App                 fbapp.App
	SignedRequestMaxAge time.Duration

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

		fileserver := http.FileServer(a.Static.Box.HTTPBox())

		mux := http.NewServeMux()
		mux.Handle(a.Static.HttpPath, a.Static)
		mux.Handle("/favicon.ico", fileserver)
		mux.Handle("/f8.jpg", fileserver)
		mux.Handle("/robots.txt", fileserver)
		mux.Handle(public, http.StripPrefix(public, fileserver))
		mux.HandleFunc("/not_a_real_webpage", http.NotFound)
		mux.Handle("/info/", a.ContextHandler)
		mux.HandleFunc("/examples/", a.ExamplesHandler.List)
		mux.HandleFunc("/saved/", a.ExamplesHandler.Saved)
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
			MaxAge:  a.SignedRequestMaxAge,
		}
		handler = httpgzip.NewHandler(handler)
		a.mainHandler = handler
	})
	a.mainHandler.ServeHTTP(w, r)
}

func (a *App) envHandler(w http.ResponseWriter, r *http.Request) {
	for _, s := range os.Environ() {
		fmt.Fprintln(w, s)
	}
}
