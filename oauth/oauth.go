// Package oauth implements an OAuth handler for Facebook.
package oauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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
	BrowserID     *browserid.Cookie
}

func (a *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case Path:
		a.Start(w, r)
		return
	case Path + resp:
		a.Response(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	h.WriteResponse(w, r, &h.Script{
		Inner: h.Unsafe("top.location='/'"),
	})
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
		values.Set("state", a.state(w, r))
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

func (a *Handler) Response(w http.ResponseWriter, r *http.Request) {
	c, err := a.ContextParser.FromRequest(r)
	if err != nil {
		view.Error(w, r, a.Static, err)
		return
	}
	if r.FormValue("state") != a.state(w, r) {
		view.Error(w, r, a.Static, errInvalidState)
		return
	}

	values := url.Values{}
	values.Set("client_id", strconv.FormatUint(a.App.ID(), 10))
	values.Set("client_secret", a.App.Secret())
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
		view.Error(w, r, a.Static, errOAuthFail)
	}
	res, err := a.HttpTransport.RoundTrip(req)
	if err != nil {
		log.Printf("oauth.Response error: %s", err)
		view.Error(w, r, a.Static, errOAuthFail)
	}
	defer res.Body.Close()
	bd, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("oauth.Response error: %s", err)
		view.Error(w, r, a.Static, errOAuthFail)
	}
	h.WriteResponse(w, r, &h.Frag{
		&h.Script{Inner: h.Unsafe("window.location.hash = ''")},
		h.String(string(bd)),
	})
}

func (a *Handler) state(w http.ResponseWriter, r *http.Request) string {
	return a.BrowserID.Get(w, r)[:10]
}

func redirectURI(c *context.Context) string {
	return c.AbsoluteURL(Path + resp).String()
}
