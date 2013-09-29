// Package oauth implements an OAuth handler for Facebook.
package oauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/daaku/go.browserid"
	"github.com/daaku/go.fbapp"
	"github.com/daaku/go.fburl"
	"github.com/daaku/go.h"
	"github.com/daaku/go.static"

	"github.com/daaku/rell/context"
	"github.com/daaku/rell/view"
)

const (
	Path = "/oauth/"
	resp = "response/"
)

var (
	errOAuthFail    = errors.New("OAuth code exchange failure.")
	errInvalidState = errors.New("Invalid state")
)

type Handler struct {
	ContextParser *context.Parser
	HttpTransport http.RoundTripper
	Static        *static.Handler
	App           fbapp.App
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case Path:
		h.Start(w, r)
		return
	case Path + resp:
		h.Response(w, r)
		return
	}
	http.NotFound(w, r)
}

func (a *Handler) Start(w http.ResponseWriter, r *http.Request) {
	c, err := a.ContextParser.FromRequest(r)
	if err != nil {
		view.Error(w, r, a.Static, err)
		return
	}
	values := url.Values{}
	values.Set("client_id", strconv.FormatUint(c.AppID, 10))
	if scope := r.FormValue("scope"); scope != "" {
		values.Set("scope", scope)
	}

	if c.ViewMode == context.Website {
		values.Set("redirect_uri", redirectURI(c))
		values.Set("state", state(w, r))
	} else {
		values.Set("redirect_uri", c.ViewURL("/auth/session"))
	}

	dialogURL := fburl.URL{
		Scheme:    "https",
		SubDomain: fburl.DWww,
		Env:       c.Env,
		Path:      "/dialog/oauth",
		Values:    values,
	}

	if c.ViewMode == context.Website {
		http.Redirect(w, r, dialogURL.String(), 302)
	} else {
		b, _ := json.Marshal(dialogURL.String())
		h.WriteResponse(w, r, &h.Script{
			Inner: h.Unsafe(fmt.Sprintf("top.location=%s", b)),
		})
	}
}

func (h *Handler) Response(w http.ResponseWriter, r *http.Request) {
	c, err := h.ContextParser.FromRequest(r)
	if err != nil {
		view.Error(w, r, h.Static, err)
		return
	}
	if r.FormValue("state") != state(w, r) {
		view.Error(w, r, h.Static, errInvalidState)
		return
	}

	values := url.Values{}
	values.Set("client_id", strconv.FormatUint(h.App.ID(), 10))
	values.Set("client_secret", h.App.Secret())
	values.Set("redirect_uri", redirectURI(c))
	values.Set("code", r.FormValue("code"))

	atURL := &fburl.URL{
		Scheme:    "https",
		SubDomain: fburl.DGraph,
		Env:       c.Env,
		Path:      "/oauth/access_token",
		Values:    values,
	}

	req, err := http.NewRequest("GET", atURL.String(), nil)
	if err != nil {
		log.Printf("oauth.Response error: %s", err)
		view.Error(w, r, h.Static, errOAuthFail)
	}
	res, err := h.HttpTransport.RoundTrip(req)
	if err != nil {
		log.Printf("oauth.Response error: %s", err)
		view.Error(w, r, h.Static, errOAuthFail)
	}
	defer res.Body.Close()
	if _, err := io.Copy(w, res.Body); err != nil {
		view.Error(w, r, h.Static, err)
		return
	}
}

func state(w http.ResponseWriter, r *http.Request) string {
	return browserid.Get(w, r)[:10]
}

func redirectURI(c *context.Context) string {
	return c.AbsoluteURL(Path + resp).String()
}
