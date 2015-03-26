// Package web provides the main HTTP entrypoint for rell.
package web

import (
	"log"
	"net/http"
	"path"
	"sync"
	"time"

	"github.com/daaku/ctxmux"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.httpgzip"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.signedrequest/appdata"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.static"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/facebookgo/fbapp"
	"github.com/daaku/rell/adminweb"
	"github.com/daaku/rell/context"
	"github.com/daaku/rell/context/viewcontext"
	"github.com/daaku/rell/examples/viewexamples"
	"github.com/daaku/rell/oauth"
	"github.com/daaku/rell/og/viewog"
	"github.com/daaku/rell/view"
	netcontext "golang.org/x/net/context"
)

// The rell web application.
type Handler struct {
	Logger              *log.Logger
	App                 fbapp.App
	SignedRequestMaxAge time.Duration
	ContextParser       *context.Parser

	ContextHandler  *viewcontext.Handler
	ExamplesHandler *viewexamples.Handler
	OgHandler       *viewog.Handler
	OauthHandler    *oauth.Handler
	Static          *static.Handler
	AdminHandler    *adminweb.Handler

	mux  http.Handler
	once sync.Once
}

// Serve HTTP requests for the main port.
func (a *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.once.Do(func() {
		const public = "/public/"

		fileserver := http.FileServer(a.Static.Box.HTTPBox())

		mux := ctxmux.Mux{
			ErrorHandler: a.handleError,
		}
		mux.Handler("GET", a.Static.HttpPath+"*rest", a.Static)
		mux.Handler("GET", "/favicon.ico", fileserver)
		mux.Handler("GET", "/f8.jpg", fileserver)
		mux.Handler("GET", "/robots.txt", fileserver)
		mux.Handler("GET", public+"*rest", http.StripPrefix(public, fileserver))
		mux.Handler("GET", "/info/*rest", a.ContextHandler)
		mux.Handler("POST", "/info/*rest", a.ContextHandler)
		mux.HandlerFunc("GET", "/examples/", a.ExamplesHandler.List)
		mux.GET("/saved/:hash", a.ExamplesHandler.GetSaved)
		mux.POST("/saved/", a.ExamplesHandler.PostSaved)
		mux.HandlerFunc("GET", "/", a.ExamplesHandler.Example)
		mux.HandlerFunc("GET", "/og/*rest", a.OgHandler.Values)
		mux.HandlerFunc("GET", "/rog/*rest", a.OgHandler.Base64)
		mux.HandlerFunc("GET", "/rog-redirect/*rest", a.OgHandler.Redirect)
		mux.Handler("GET", oauth.Path+"*rest", a.OauthHandler)

		if a.AdminHandler.Path != "" {
			mux.Handler("GET", path.Join("/", a.AdminHandler.Path)+"/", a.AdminHandler)
		}

		var handler http.Handler
		handler = &appdata.Handler{
			Handler: &mux,
			Secret:  a.App.SecretByte(),
			MaxAge:  a.SignedRequestMaxAge,
		}
		handler = httpgzip.NewHandler(handler)
		a.mux = handler
	})
	a.mux.ServeHTTP(w, r)
}

func (a *Handler) handleError(ctx netcontext.Context, w http.ResponseWriter, r *http.Request, err error) {
	a.Logger.Println(err)
	view.Error(w, r, a.Static, err)
}
