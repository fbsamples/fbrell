// Package oauth implements an OAuth handler for Facebook.
package oauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.browserid"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.fburl"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.h"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.static"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/facebookgo/fbapp"
	"github.com/daaku/rell/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/daaku/rell/rellenv"
)

const (
	Path = "/oauth/"
	resp = "response/"
)

var (
	errOAuthFail     = errors.New("OAuth code exchange failure.")
	errInvalidState  = errors.New("Invalid state")
	errEmployeesOnly = errors.New("This endpoint is for employees only.")
)

type Handler struct {
	HttpTransport http.RoundTripper
	Static        *static.Handler
	App           fbapp.App
	BrowserID     *browserid.Cookie
}

func (a *Handler) Handler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	c, err := rellenv.FromContext(ctx)
	if err != nil {
		return err
	}
	if !c.IsEmployee {
		return errEmployeesOnly
	}

	switch r.URL.Path {
	case Path:
		return a.Start(ctx, w, r)
	case Path + resp:
		return a.Response(ctx, w, r)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	h.WriteResponse(w, r, &h.Script{
		Inner: h.Unsafe("top.location='/'"),
	})
	return nil
}

func (a *Handler) Start(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	c, err := rellenv.FromContext(ctx)
	if err != nil {
		return err
	}
	values := url.Values{}
	values.Set("client_id", strconv.FormatUint(c.AppID, 10))
	if scope := r.FormValue("scope"); scope != "" {
		values.Set("scope", scope)
	}

	if c.ViewMode == rellenv.Website {
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

	if c.ViewMode == rellenv.Website {
		http.Redirect(w, r, dialogURL.String(), 302)
	} else {
		b, _ := json.Marshal(dialogURL.String())
		h.WriteResponse(w, r, &h.Script{
			Inner: h.Unsafe(fmt.Sprintf("top.location=%s", b)),
		})
	}
	return nil
}

func (a *Handler) Response(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	c, err := rellenv.FromContext(ctx)
	if err != nil {
		return err
	}
	if r.FormValue("state") != a.state(w, r) {
		return errInvalidState
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
		return errOAuthFail
	}
	res, err := a.HttpTransport.RoundTrip(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	bd, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	h.WriteResponse(w, r, &h.Frag{
		&h.Script{Inner: h.Unsafe("window.location.hash = ''")},
		h.String(string(bd)),
	})
	return nil
}

func (a *Handler) state(w http.ResponseWriter, r *http.Request) string {
	return a.BrowserID.Get(w, r)[:10]
}

func redirectURI(c *rellenv.Env) string {
	return c.AbsoluteURL(Path + resp).String()
}
