// Package web provides the main HTTP entrypoint for rell.
package web

import (
	"net/http"
	"path"
	"sync"
	"time"

	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.httpgzip"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.signedrequest/appdata"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.static"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/facebookgo/fbapp"
	"github.com/daaku/rell/adminweb"
	"github.com/daaku/rell/context/viewcontext"
	"github.com/daaku/rell/examples/viewexamples"
	"github.com/daaku/rell/oauth"
	"github.com/daaku/rell/og/viewog"
)

// The rell web application.
type Handler struct {
	ContextHandler      *viewcontext.Handler
	ExamplesHandler     *viewexamples.Handler
	OgHandler           *viewog.Handler
	OauthHandler        *oauth.Handler
	Static              *static.Handler
	AdminHandler        *adminweb.Handler
	App                 fbapp.App
	SignedRequestMaxAge time.Duration

	mux  http.Handler
	once sync.Once
}

// Serve HTTP requests for the main port.
func (a *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.once.Do(func() {
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

		if a.AdminHandler.Path != "" {
			mux.Handle(path.Join("/", a.AdminHandler.Path)+"/", a.AdminHandler)
		}

		var handler http.Handler
		handler = &appdata.Handler{
			Handler: mux,
			Secret:  a.App.SecretByte(),
			MaxAge:  a.SignedRequestMaxAge,
		}
		handler = httpgzip.NewHandler(handler)
		a.mux = handler
	})
	a.mux.ServeHTTP(w, r)
}
