// Package oauth implements an OAuth handler for Facebook.
package oauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/daaku/go.browserid"
	"github.com/daaku/go.fbapp"
	"github.com/daaku/go.fburl"
	"github.com/daaku/go.h"

	"github.com/daaku/rell/context"
	"github.com/daaku/rell/service"
	"github.com/daaku/rell/view"
)

const (
	Path = "/oauth/"
	resp = "response/"
)

var errInvalidState = errors.New("Invalid state")

func Handle(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case Path:
		Start(w, r)
		return
	case Path + resp:
		Response(w, r)
		return
	}
	http.NotFound(w, r)
}

func Start(w http.ResponseWriter, r *http.Request) {
	c, err := context.FromRequest(r)
	if err != nil {
		view.Error(w, r, err)
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

func Response(w http.ResponseWriter, r *http.Request) {
	c, err := context.FromRequest(r)
	if err != nil {
		view.Error(w, r, err)
		return
	}
	if r.FormValue("state") != state(w, r) {
		view.Error(w, r, errInvalidState)
		return
	}
	values := url.Values{}
	values.Set("client_id", strconv.FormatUint(fbapp.Default.ID(), 10))
	values.Set("client_secret", fbapp.Default.Secret())
	values.Set("redirect_uri", redirectURI(c))
	values.Set("code", r.FormValue("code"))
	res, err := service.FbApiClient.GetRaw("/oauth/access_token", values)
	if err != nil {
		view.Error(w, r, err)
		return
	}
	w.Write(res)
}

func state(w http.ResponseWriter, r *http.Request) string {
	return browserid.Get(w, r)[:10]
}

func redirectURI(c *context.Context) string {
	return c.AbsoluteURL(Path + resp).String()
}
