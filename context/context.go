// Package context implements the shared context for a Rell
// request, including the parsed global state associated with URLs and
// the SDK version.
package context

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"

	"github.com/daaku/go.fburl"
	"github.com/daaku/go.signedrequest/appdata"
	"github.com/daaku/go.signedrequest/fbsr"
	"github.com/daaku/go.stats"
	"github.com/daaku/go.trustforward"
	"github.com/facebookgo/fbapp"
	"github.com/gorilla/schema"

	"github.com/daaku/rell/context/appns"
	"github.com/daaku/rell/context/empcheck"
)

var envRegexp = regexp.MustCompile(`^[a-zA-Z0-9-_.]*$`)

const defaultMaxMemory = 32 << 20 // 32 MB

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
	defaultAppID         uint64              `schema:"-"`
	AppNamespace         string              `schema:"-"`
	Level                string              `schema:"level"`
	Locale               string              `schema:"locale"`
	Env                  string              `schema:"server"`
	Trace                bool                `schema:"trace"`
	Status               bool                `schema:"status"`
	FrictionlessRequests bool                `schema:"frictionlessRequests"`
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
	Status:               true,
	FrictionlessRequests: true,
	Host:                 "www.fbrell.com",
	Scheme:               "http",
	ViewMode:             Website,
	ViewportMode:         ViewportModeMobile,
	Module:               "all",
	Init:                 true,
}

var (
	schemaDecoder = schema.NewDecoder()
)

type Parser struct {
	EmpChecker   *empcheck.Checker
	AppNSFetcher *appns.Fetcher
	App          fbapp.App
	Stats        stats.Backend
}

// Create a default context.
func (p *Parser) Default() *Context {
	context := defaultContext.Copy()
	context.AppID = p.App.ID()
	context.defaultAppID = p.App.ID()
	return context
}

// Create a context from a HTTP request.
func (p *Parser) FromRequest(r *http.Request) (*Context, error) {
	err := r.ParseMultipartForm(defaultMaxMemory)
	if err != nil && err != http.ErrNotMultipart {
		return nil, err
	}
	if id := r.FormValue("client_id"); id != "" {
		r.Form.Set("appid", id)
	}
	context := p.Default()
	_ = schemaDecoder.Decode(context, r.URL.Query())
	_ = schemaDecoder.Decode(context, r.Form)
	rawSr := r.FormValue("signed_request")
	if rawSr != "" {
		context.SignedRequest, err = fbsr.Unmarshal(
			[]byte(rawSr), p.App.SecretByte())
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
				[]byte(cookie.Value), p.App.SecretByte())
		}
	}
	context.Host = trustforward.Host(r)
	context.Scheme = trustforward.Scheme(r)
	if context.SignedRequest != nil && context.SignedRequest.UserID != 0 {
		context.IsEmployee = p.EmpChecker.Check(context.SignedRequest.UserID)
	}
	context.AppNamespace = p.AppNSFetcher.Get(context.AppID)
	if context.Env != "" && !envRegexp.MatchString(context.Env) {
		context.Env = ""
	}
	return context, nil
}

// Provides a duplicate copy.
func (c *Context) Copy() *Context {
	context := *c
	return &context
}

// Get the URL for the JS SDK.
func (c *Context) SdkURL() string {
	server := "connect.facebook.net"
	if c.Env != "" {
		server = fburl.Hostname("static", c.Env) + "/assets.php"
	}
	return fmt.Sprintf("%s://%s/%s/%s.js", c.Scheme, server, c.Locale, c.Module)
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
	var base = "/" + c.AppNamespace + "/"
	if name == "" || name == "/" {
		name = base
	} else {
		name = path.Join(base, name)
	}

	url := fburl.URL{
		Scheme:    "https",
		SubDomain: fburl.DApps,
		Env:       c.Env,
		Path:      name,
		Values:    c.Values(),
	}
	return url.String()
}

// Serialize the context back to URL values.
func (c *Context) Values() url.Values {
	values := url.Values{}
	if c.AppID != c.defaultAppID {
		values.Set("appid", strconv.FormatUint(c.AppID, 10))
	}
	if c.Env != defaultContext.Env {
		values.Set("server", c.Env)
	}
	if c.Locale != defaultContext.Locale {
		values.Set("locale", c.Locale)
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
		"status":               c.Status,
		"frictionlessRequests": c.FrictionlessRequests,
		"signedRequest":        c.SignedRequest,
		"viewMode":             c.ViewMode,
		"init":                 c.Init,
	}
	if c.IsEmployee {
		data["isEmployee"] = true
	}
	return json.Marshal(data)
}
