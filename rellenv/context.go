// Package context implements the shared context for a Rell
// request, including the parsed global state associated with URLs and
// the SDK version.
package rellenv

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"time"

	"github.com/daaku/ctxerr"
	"github.com/daaku/go.fburl"
	"github.com/daaku/go.signedrequest/appdata"
	"github.com/daaku/go.signedrequest/fbsr"
	"github.com/daaku/go.trustforward"
	"github.com/facebookgo/fbapp"
	"golang.org/x/net/context"
)

var envRegexp = regexp.MustCompile(`^[a-zA-Z0-9-_.]*$`)

const (
	// View Modes.
	Website = "website"
	Canvas  = "canvas"
	PageTab = "page-tab"
)

// The Context defined by the environment and as configured by the
// user via the URL.
type Env struct {
	appID                uint64
	defaultAppID         uint64
	appNamespace         string
	level                string
	locale               string
	Env                  string
	Status               bool
	FrictionlessRequests bool
	Host                 string
	Scheme               string
	SignedRequest        *fbsr.SignedRequest
	ViewMode             string
	Module               string
	isEmployee           bool
	Init                 bool
}

// Defaults for the context.
var defaultContext = &Env{
	level:                "debug",
	locale:               "en_US",
	Status:               true,
	FrictionlessRequests: true,
	Host:                 "www.fbrell.com",
	Scheme:               "http",
	ViewMode:             Website,
	Module:               "sdk",
	Init:                 true,
}

type EmpChecker interface {
	Check(uid uint64) bool
}

type AppNSFetcher interface {
	Get(id uint64) string
}

type Parser struct {
	EmpChecker          EmpChecker
	AppNSFetcher        AppNSFetcher
	App                 fbapp.App
	SignedRequestMaxAge time.Duration
	Forwarded           *trustforward.Forwarded
}

// Create a default context.
func (p *Parser) Default() *Env {
	context := defaultContext.Copy()
	context.appID = p.App.ID()
	context.defaultAppID = p.App.ID()
	return context
}

// Create a context from a HTTP request.
func (p *Parser) FromRequest(ctx context.Context, r *http.Request) (*Env, error) {
	e := p.Default()

	if appid, err := strconv.ParseUint(r.FormValue("appid"), 10, 64); err == nil {
		e.appID = appid
	}
	if appid, err := strconv.ParseUint(r.FormValue("client_id"), 10, 64); err == nil {
		e.appID = appid
	}
	if level := r.FormValue("level"); level != "" {
		e.level = level
	}
	if locale := r.FormValue("locale"); locale != "" {
		e.locale = locale
	}
	if env := r.FormValue("server"); env != "" {
		e.Env = env
	}
	if viewMode := r.FormValue("view-mode"); viewMode != "" {
		e.ViewMode = viewMode
	}
	if module := r.FormValue("module"); module != "" {
		e.Module = module
	}
	if status, err := strconv.ParseBool(r.FormValue("status")); err == nil {
		e.Status = status
	}
	if fr, err := strconv.ParseBool(r.FormValue("frictionlessRequests")); err == nil {
		e.FrictionlessRequests = fr
	}
	if init, err := strconv.ParseBool(r.FormValue("init")); err == nil {
		e.Init = init
	}

	var err error
	rawSr := r.FormValue("signed_request")
	if rawSr != "" {
		e.SignedRequest, err = fbsr.Unmarshal(
			[]byte(rawSr),
			p.App.SecretByte(),
			p.SignedRequestMaxAge,
		)
		if err == nil {
			if e.SignedRequest.Page != nil {
				e.ViewMode = PageTab
			} else {
				e.ViewMode = Canvas
			}
		}
	} else {
		cookie, _ := r.Cookie(fmt.Sprintf("fbsr_%d", e.appID))
		if cookie != nil {
			e.SignedRequest, err = fbsr.Unmarshal(
				[]byte(cookie.Value),
				p.App.SecretByte(),
				p.SignedRequestMaxAge,
			)
		}
	}
	e.Host = p.Forwarded.Host(r)
	e.Scheme = p.Forwarded.Scheme(r)
	if e.SignedRequest != nil && e.SignedRequest.UserID != 0 {
		e.isEmployee = p.EmpChecker.Check(e.SignedRequest.UserID)
	}
	e.appNamespace = p.AppNSFetcher.Get(e.appID)
	if e.Env != "" && !envRegexp.MatchString(e.Env) {
		e.Env = ""
	}
	return e, nil
}

// Provides a duplicate copy.
func (c *Env) Copy() *Env {
	context := *c
	return &context
}

// Get the URL for the JS SDK.
func (c *Env) SdkURL() string {
	server := "connect.facebook.net"
	if c.Env != "" {
		server = fburl.Hostname("static", c.Env) + "/assets.php"
	}
	return fmt.Sprintf("%s://%s/%s/%s.js", c.Scheme, server, c.locale, c.Module)
}

// Get the URL for loading this application in a Page Tab on Facebook.
func (c *Env) PageTabURL(name string) string {
	values := url.Values{}
	values.Set("sk", fmt.Sprintf("app_%d", c.appID))
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
func (c *Env) CanvasURL(name string) string {
	var base = "/" + c.appNamespace + "/"
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
func (c *Env) Values() url.Values {
	values := url.Values{}
	if c.appID != c.defaultAppID {
		values.Set("appid", strconv.FormatUint(c.appID, 10))
	}
	if c.Env != defaultContext.Env {
		values.Set("server", c.Env)
	}
	if c.locale != defaultContext.locale {
		values.Set("locale", c.locale)
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
func (c *Env) URL(path string) *url.URL {
	return &url.URL{
		Path:     path,
		RawQuery: c.Values().Encode(),
	}
}

// Create a context aware absolute URL for the given path.
func (c *Env) AbsoluteURL(path string) *url.URL {
	u := c.URL(path)
	u.Host = c.Host
	u.Scheme = c.Scheme
	return u
}

// This will return a view aware URL and will always be absolute.
func (c *Env) ViewURL(path string) string {
	switch c.ViewMode {
	case Canvas:
		return c.CanvasURL(path)
	case PageTab:
		return c.PageTabURL(path)
	default:
		return c.AbsoluteURL(path).String()
	}
}

// JSON representation of Context.
func (c *Env) MarshalJSON() ([]byte, error) {
	data := map[string]interface{}{
		"appID":                strconv.FormatUint(c.appID, 10),
		"level":                c.level,
		"status":               c.Status,
		"frictionlessRequests": c.FrictionlessRequests,
		"signedRequest":        c.SignedRequest,
		"viewMode":             c.ViewMode,
		"init":                 c.Init,
	}
	if c.isEmployee {
		data["isEmployee"] = true
	}
	return json.Marshal(data)
}

type contextEnvKeyT int

var contextEnvKey = contextEnvKeyT(1)

var errEnvNotFound = errors.New("rellenv: Env not found in Context")

// FromContext retrieves the Env from the Context. If one isn't found, an error
// is returned.
func FromContext(ctx context.Context) (*Env, error) {
	if e, ok := ctx.Value(contextEnvKey).(*Env); ok {
		return e, nil
	}
	return nil, ctxerr.Wrap(ctx, errEnvNotFound)
}

// WithEnv adds the given env to the context.
func WithEnv(ctx context.Context, env *Env) context.Context {
	return context.WithValue(ctx, contextEnvKey, env)
}

// IsEmployee returns true if the Context is known to be that of an employee.
func IsEmployee(ctx context.Context) bool {
	if ctx, err := FromContext(ctx); err == nil {
		return ctx.isEmployee
	}
	return false
}

var defaultFbApp = fbapp.New(342526215814610, "", "")

// FbApp returns the FB application configured in the context. Generally this
// doesn't make sense as something that's different per context, but for FB
// employees this is a meta-tool of sorts and this makes complex things
// possible.
func FbApp(ctx context.Context) fbapp.App {
	// TODO: secret?
	if ctx, err := FromContext(ctx); err == nil {
		return fbapp.New(ctx.appID, "", ctx.appNamespace)
	}
	return defaultFbApp
}

func FbEnv(ctx context.Context) string {
	if ctx, err := FromContext(ctx); err == nil {
		return ctx.Env
	}
	return ""
}
