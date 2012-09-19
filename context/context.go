// Package context implements the shared context for a Rell
// request, including the parsed global state associated with URLs and
// the SDK version.
package context

import (
	"code.google.com/p/gorilla/schema"
	"encoding/json"
	"fmt"
	"github.com/daaku/go.fbapi"
	"github.com/daaku/go.fbapp"
	"github.com/daaku/go.fburl"
	"github.com/daaku/go.signedrequest/appdata"
	"github.com/daaku/go.signedrequest/fbsr"
	"github.com/daaku/go.stats"
	"github.com/daaku/go.trustforward"
	"github.com/daaku/rell/context/empcheck"
	"net/http"
	"net/url"
	"path"
	"strconv"
)

const defaultMaxMemory = 32 << 20 // 32 MB

// The allowed SDK Versions.
const (
	Mu  = "mu"
	Old = "old"
	Mid = "mid"
)

const (
	// View Modes.
	Website = "website"
	Canvas  = "canvas"
	PageTab = "page-tab"

	// View Port
	ViewportModeMobile = "mobile"
	ViewportModeAuto   = ""
)

// The Context defined by the environment and as configured by the
// user via the URL.
type Context struct {
	AppID                uint64              `schema:"appid"`
	Level                string              `schema:"level"`
	Locale               string              `schema:"locale"`
	Env                  string              `schema:"server"`
	Trace                bool                `schema:"trace"`
	Version              string              `schema:"version"`
	Status               bool                `schema:"status"`
	FrictionlessRequests bool                `schema:"frictionlessRequests"`
	UseChannel           bool                `schema:"channel"`
	Host                 string              `schema:"-"`
	Scheme               string              `schema:"-"`
	SignedRequest        *fbsr.SignedRequest `schema:"-"`
	ViewMode             string              `schema:"view-mode"`
	Module               string              `schema:"module"`
	ViewportMode         string              `schema:"viewport-mode"`
	IsEmployee           bool                `schema:"-"`
	Init                 bool                `schema:"init"`
}

// Defaults for the context.
var defaultContext = &Context{
	Level:                "debug",
	Locale:               "en_US",
	Version:              Mu,
	Status:               true,
	FrictionlessRequests: true,
	UseChannel:           true,
	Host:                 "www.fbrell.com",
	Scheme:               "http",
	ViewMode:             Website,
	ViewportMode:         ViewportModeMobile,
	Module:               "all",
	Init:                 true,
}

var schemaDecoder = schema.NewDecoder()

// Create a context from a HTTP request.
func FromRequest(r *http.Request) (*Context, error) {
	err := r.ParseMultipartForm(defaultMaxMemory)
	if err != nil {
		return nil, err
	}
	if id := r.FormValue("client_id"); id != "" {
		r.Form.Set("appid", id)
	}
	context := Default()
	_ = schemaDecoder.Decode(context, r.URL.Query())
	_ = schemaDecoder.Decode(context, r.Form)
	rawSr := r.FormValue("signed_request")
	if rawSr != "" {
		context.SignedRequest, err = fbsr.Unmarshal(
			[]byte(rawSr), fbapp.Default.SecretByte())
		if err == nil {
			if context.SignedRequest.Page != nil {
				context.ViewMode = PageTab
			} else {
				context.ViewMode = Canvas
			}
		}
	} else {
		cookie, _ := r.Cookie(fmt.Sprintf("fbsr_%d", context.AppID))
		if cookie != nil {
			context.SignedRequest, err = fbsr.Unmarshal(
				[]byte(cookie.Value), fbapp.Default.SecretByte())
		}
	}
	context.Host = trustforward.Host(r)
	context.Scheme = trustforward.Scheme(r)
	if context.SignedRequest != nil && context.SignedRequest.UserID != 0 {
		context.IsEmployee = empcheck.IsEmployee(context.SignedRequest.UserID)
	}
	return context, nil
}

// Provides a duplicate copy.
func (c *Context) Copy() *Context {
	context := *c
	return &context
}

// Create a default context.
func Default() *Context {
	context := defaultContext.Copy()
	context.AppID = fbapp.Default.ID()
	return context
}

// Get the URL for the JS SDK.
func (c *Context) SdkURL() string {
	var server string
	if c.Version == Mu {
		if c.Env == "" {
			server = "connect.facebook.net"
		} else {
			server = fburl.Hostname("static", c.Env) + "/assets.php"
		}
		return fmt.Sprintf("%s://%s/%s/%s.js", c.Scheme, server, c.Locale, c.Module)
	} else {
		if c.Env == "" {
			if c.Scheme == "https" {
				server = "s-static.ak.facebook.com"
			} else {
				server = "static.ak.facebook.com"
			}
		} else {
			server = fburl.Hostname("static", c.Env)
		}
		if c.Version == Mid {
			return fmt.Sprintf(
				"%s://%s/connect.php/%s/js", c.Scheme, server, c.Locale)
		} else {
			return fmt.Sprintf(
				"%s://%s/js/api_lib/v0.4/FeatureLoader.js.php", c.Scheme, server)
		}
	}
	panic("Not reached")
}

// Get the URL for loading this application in a Page Tab on Facebook.
func (c *Context) PageTabURL(name string) string {
	values := url.Values{}
	values.Set("sk", fmt.Sprintf("app_%d", c.AppID))
	values.Set("app_data", appdata.Encode(c.URL(name)))
	url := fburl.URL{
		Scheme:    c.Scheme,
		SubDomain: fburl.DWww,
		Env:       c.Env,
		Path:      "/pages/Rell-Page-for-Tabs/141929622497380",
		Values:    values,
	}
	return url.String()
}

// Get the URL for loading this application in a Canvas page on Facebook.
func (c *Context) CanvasURL(name string) string {
	var base = "/" + c.AppNamespace() + "/"
	if name == "" || name == "/" {
		name = base
	} else {
		name = path.Join(base, name)
	}

	url := fburl.URL{
		Scheme:    c.Scheme,
		SubDomain: fburl.DApps,
		Env:       c.Env,
		Path:      name,
		Values:    c.Values(),
	}
	return url.String()
}

// Get the App Namespace, fetching it using the Graph API if necessary.
func (c *Context) AppNamespace() string {
	if c.AppID == fbapp.Default.ID() {
		return fbapp.Default.Namespace()
	}
	stats.Inc("context app namespace fetch")
	resp := struct{ Namespace string }{""}
	err := fbapi.Get(&resp, fmt.Sprintf("/%d", c.AppID), fbapi.Fields{"namespace"})
	if err != nil {
		stats.Inc("context app namespace fetch failure")
	}
	return resp.Namespace
}

// Get a Channel URL for the SDK.
func (c *Context) ChannelURL() string {
	return c.AbsoluteURL("/channel/").String()
}

// Serialize the context back to URL values.
func (c *Context) Values() url.Values {
	values := url.Values{}
	if c.AppID != fbapp.Default.ID() {
		values.Set("appid", strconv.FormatUint(c.AppID, 10))
	}
	if c.Env != defaultContext.Env {
		values.Set("server", c.Env)
	}
	if c.Locale != defaultContext.Locale {
		values.Set("locale", c.Locale)
	}
	if c.Version != defaultContext.Version {
		values.Set("version", c.Version)
	}
	if c.ViewportMode != defaultContext.ViewportMode {
		values.Set("viewport-mode", c.ViewportMode)
	}
	if c.Module != defaultContext.Module {
		values.Set("module", c.Module)
	}
	if c.Init != defaultContext.Init {
		values.Set("init", strconv.FormatBool(c.Init))
	}
	if c.Status != defaultContext.Status {
		values.Set("status", strconv.FormatBool(c.Status))
	}
	if c.UseChannel != defaultContext.UseChannel {
		values.Set("channel", strconv.FormatBool(c.UseChannel))
	}
	if c.FrictionlessRequests != defaultContext.FrictionlessRequests {
		values.Set("frictionlessRequests", strconv.FormatBool(c.FrictionlessRequests))
	}
	return values
}

// Create a context aware URL for the given path.
func (c *Context) URL(path string) *url.URL {
	return &url.URL{
		Path:     path,
		RawQuery: c.Values().Encode(),
	}
}

// Create a context aware absolute URL for the given path.
func (c *Context) AbsoluteURL(path string) *url.URL {
	u := c.URL(path)
	u.Host = c.Host
	u.Scheme = c.Scheme
	return u
}

// This will return a view aware URL and will always be absolute.
func (c *Context) ViewURL(path string) string {
	switch c.ViewMode {
	case Canvas:
		return c.CanvasURL(path)
	case PageTab:
		return c.PageTabURL(path)
	default:
		return c.AbsoluteURL(path).String()
	}
	panic("not reached")
}

// Context aware viewport for a customized mobile experience.
func (c *Context) Viewport() string {
	if c.ViewportMode == ViewportModeMobile {
		return "width=device-width,initial-scale=1.0"
	}
	return ""
}

// JSON representation of Context.
func (c *Context) MarshalJSON() ([]byte, error) {
	data := map[string]interface{}{
		"appID":                strconv.FormatUint(c.AppID, 10),
		"level":                c.Level,
		"trace":                c.Trace,
		"version":              c.Version,
		"status":               c.Status,
		"frictionlessRequests": c.FrictionlessRequests,
		"channel":              c.UseChannel,
		"channelURL":           c.ChannelURL(),
		"signedRequest":        c.SignedRequest,
		"viewMode":             c.ViewMode,
		"init":                 c.Init,
	}
	if c.IsEmployee {
		data["isEmployee"] = true
	}
	return json.Marshal(data)
}
