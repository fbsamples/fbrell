/**
 * Copyright (c) 2014-present, Facebook, Inc. All rights reserved.
 *
 * You are hereby granted a non-exclusive, worldwide, royalty-free license to use,
 * copy, modify, and distribute this software in source code or binary form for use
 * in connection with the web services and APIs provided by Facebook.
 *
 * As with any software that integrates with the Facebook platform, your use of
 * this software is subject to the Facebook Developer Principles and Policies
 * [http://developers.facebook.com/policy/]. This copyright notice shall be
 * included in all copies or substantial portions of the software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

// Package oauth implements an OAuth handler for Facebook.
package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/daaku/ctxerr"
	"github.com/daaku/go.browserid"
	"github.com/daaku/go.fburl"
	"github.com/daaku/go.h"
	"github.com/daaku/go.static"
	"github.com/facebookgo/fbapp"
	"github.com/fbsamples/fbrell/rellenv"
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

func (a *Handler) Handler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	if !rellenv.IsEmployee(ctx) {
		return ctxerr.Wrap(ctx, errEmployeesOnly)
	}

	switch r.URL.Path {
	case Path:
		return a.Start(ctx, w, r)
	case Path + resp:
		return a.Response(ctx, w, r)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	_, err := h.Write(ctx, w, &h.Script{
		Inner: h.Unsafe("top.location='/'"),
	})
	return err
}

func (a *Handler) Start(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	c, err := rellenv.FromContext(ctx)
	if err != nil {
		return err
	}
	values := url.Values{}
	values.Set("client_id", strconv.FormatUint(rellenv.FbApp(ctx).ID(), 10))
	if scope := r.FormValue("scope"); scope != "" {
		values.Set("scope", scope)
	}

	if assetScope := r.FormValue("asset-scope"); assetScope != "" {
		values.Set("asset-scope", assetScope)
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
		Env:       rellenv.FbEnv(ctx),
		Path:      "/dialog/oauth",
		Values:    values,
	}

	if c.ViewMode == rellenv.Website {
		http.Redirect(w, r, dialogURL.String(), 302)
	} else {
		b, _ := json.Marshal(dialogURL.String())
		_, err := h.Write(ctx, w, &h.Script{
			Inner: h.Unsafe(fmt.Sprintf("top.location=%s", b)),
		})
		return err
	}
	return nil
}

func (a *Handler) Response(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	c, err := rellenv.FromContext(ctx)
	if err != nil {
		return err
	}
	if r.FormValue("state") != a.state(w, r) {
		return ctxerr.Wrap(ctx, errInvalidState)
	}

	values := url.Values{}
	values.Set("client_id", strconv.FormatUint(a.App.ID(), 10))
	values.Set("client_secret", a.App.Secret())
	values.Set("redirect_uri", redirectURI(c))
	values.Set("code", r.FormValue("code"))

	atURL := &fburl.URL{
		Scheme:    "https",
		SubDomain: fburl.DGraph,
		Env:       rellenv.FbEnv(ctx),
		Path:      "/oauth/access_token",
		Values:    values,
	}

	req, err := http.NewRequest("GET", atURL.String(), nil)
	if err != nil {
		return ctxerr.Wrap(ctx, errOAuthFail)
	}
	res, err := a.HttpTransport.RoundTrip(req)
	if err != nil {
		return ctxerr.Wrap(ctx, err)
	}
	defer res.Body.Close()
	bd, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return ctxerr.Wrap(ctx, err)
	}
	_, err = h.Write(ctx, w, h.Frag{
		&h.Script{Inner: h.Unsafe("window.location.hash = ''")},
		h.String(string(bd)),
	})
	return err
}

func (a *Handler) state(w http.ResponseWriter, r *http.Request) string {
	return a.BrowserID.Get(w, r)[:10]
}

func redirectURI(c *rellenv.Env) string {
	return c.AbsoluteURL(Path + resp).String()
}
