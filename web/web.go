// Package web provides the main HTTP entrypoint for rell.
package web

import (
	"log"
	"net/http"
	"path"
	"sync"
	"time"

	"github.com/daaku/rell/adminweb"
	"github.com/daaku/rell/examples/viewexamples"
	"github.com/daaku/rell/internal/github.com/daaku/ctxerr"
	"github.com/daaku/rell/internal/github.com/daaku/ctxmux"
	"github.com/daaku/rell/internal/github.com/daaku/go.signedrequest/appdata"
	"github.com/daaku/rell/internal/github.com/daaku/go.static"
	"github.com/daaku/rell/internal/github.com/facebookgo/fbapp"
	"github.com/daaku/rell/internal/golang.org/x/net/context"
	"github.com/daaku/rell/oauth"
	"github.com/daaku/rell/og/viewog"
	"github.com/daaku/rell/rellenv"
	"github.com/daaku/rell/rellenv/viewcontext"
	"github.com/daaku/rell/view"
)

// The rell web application.
type Handler struct {
	Logger              *log.Logger
	App                 fbapp.App
	SignedRequestMaxAge time.Duration
	EnvParser           *rellenv.Parser
	PublicFS            http.FileSystem

	ContextHandler  *viewcontext.Handler
	ExamplesHandler *viewexamples.Handler
	OgHandler       *viewog.Handler
	OauthHandler    *oauth.Handler
	Static          *static.Handler
	AdminHandler    *adminweb.Handler

	ctx  context.Context
	mux  http.Handler
	once sync.Once
}

// Serve HTTP requests for the main port.
func (a *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.once.Do(func() {
		const public = "/public/"

		fileserver := http.FileServer(a.PublicFS)

		mux, err := ctxmux.New(
			ctxmux.MuxErrorHandler(a.handleError),
			ctxmux.MuxNotFoundHandler(a.ExamplesHandler.Example),
			ctxmux.MuxRedirectTrailingSlash(),
			ctxmux.MuxContextMaker(a.contextMaker),
		)
		if err != nil {
			panic(err)
		}

		mux.GET(a.Static.Path+"*rest", ctxmux.HTTPHandler(a.Static))
		mux.GET("/favicon.ico", ctxmux.HTTPHandler(fileserver))
		mux.GET("/f8.jpg", ctxmux.HTTPHandler(fileserver))
		mux.GET("/robots.txt", ctxmux.HTTPHandler(fileserver))
		mux.GET(public+"*rest", ctxmux.HTTPHandler(http.StripPrefix(public, fileserver)))
		mux.GET("/info/*rest", a.ContextHandler.Info)
		mux.POST("/info/*rest", a.ContextHandler.Info)
		mux.GET("/examples/", a.ExamplesHandler.List)
		mux.GET("/saved/:hash", a.ExamplesHandler.GetSaved)
		mux.POST("/saved/", a.ExamplesHandler.PostSaved)
		mux.GET("/og/*rest", a.OgHandler.Values)
		mux.GET("/rog/*rest", a.OgHandler.Base64)
		mux.GET("/rog-redirect/*rest", a.OgHandler.Redirect)
		mux.GET(oauth.Path+"*rest", a.OauthHandler.Handler)

		if a.AdminHandler.Path != "" {
			adminPath := path.Join("/", a.AdminHandler.Path) + "/*rest"
			mux.GET(adminPath, ctxmux.HTTPHandler(a.AdminHandler))
		}

		var handler http.Handler
		handler = &appdata.Handler{
			Handler: mux,
			Secret:  a.App.SecretByte(),
			MaxAge:  a.SignedRequestMaxAge,
		}
		a.mux = handler

		a.ctx = context.Background()
		a.ctx = ctxerr.WithConfig(a.ctx, ctxerr.Config{
			StackMode:  ctxerr.StackModeMultiStack,
			StringMode: ctxerr.StringModeNone,
		})
	})
	a.mux.ServeHTTP(w, r)
}

func (a *Handler) handleError(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	a.Logger.Printf("Error at %s\n%s\n", r.URL, ctxerr.RichString(err))
	view.Error(w, r, a.Static, err)
}

func (a *Handler) contextMaker(r *http.Request) (context.Context, error) {
	env, err := a.EnvParser.FromRequest(a.ctx, r)
	if err != nil {
		return a.ctx, err
	}
	return rellenv.WithEnv(a.ctx, env), nil
}
