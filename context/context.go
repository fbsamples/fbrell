// Package context implements the shared context for a Rell
// request, including the parsed global state associated with URLs and
// the SDK version.
package context

import (
	"code.google.com/p/gorilla/schema"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/nshah/go.fburl"
	"github.com/nshah/go.signedrequest/fbsr"
	"log"
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

// View Modes.
const (
	Website = "website"
	Canvas  = "canvas"
	PageTab = "page-tab"
)

type App struct {
	ID        uint64
	Secret    string
	Namespace string
}

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
}

var (
	schemaDecoder = schema.NewDecoder()
	defaultApp    = &App{}
)

func init() {
	flag.Uint64Var(
		&defaultApp.ID, "rell.client.id", 184484190795, "Default Client ID.")
	flag.StringVar(
		&defaultApp.Secret, "rell.client.secret", "", "Default Client Secret.")
	flag.StringVar(
		&defaultApp.Namespace,
		"rell.client.namespace",
		"",
		"Default Client Namespace.")
}

// Create a context from a HTTP request.
func FromRequest(r *http.Request) (*Context, error) {
	r.ParseMultipartForm(defaultMaxMemory)
	context := Default()
	_ = schemaDecoder.Decode(context, r.URL.Query())
	_ = schemaDecoder.Decode(context, r.Form)
	rawSr := r.FormValue("signed_request")
	if rawSr != "" {
		var err error
		context.SignedRequest, err = fbsr.Unmarshal(
			[]byte(rawSr), []byte(defaultApp.Secret))
		if err != nil {
			log.Printf("Ignoring error in parsing signed request: %s", err)
		} else {
			if context.SignedRequest != nil && context.SignedRequest.AppData != "" {
				context.Env = context.SignedRequest.AppData
			}
			if context.SignedRequest.Page != nil {
				context.ViewMode = PageTab
			} else {
				context.ViewMode = Canvas
			}
		}
	}
	if fwdHost := r.Header.Get("x-forwarded-host"); fwdHost != "" {
		context.Host = fwdHost
	} else {
		context.Host = r.Host
	}
	if r.Header.Get("x-forwarded-proto") == "https" {
		context.Scheme = "https"
	} else {
		context.Scheme = "http"
	}
	return context, nil
}

// Create a default context.
func Default() *Context {
	context := *defaultContext
	context.AppID = defaultApp.ID
	return &context
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
		return fmt.Sprintf("%s://%s/%s/all.js", c.Scheme, server, c.Locale)
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
	ctxValues := c.Values()
	if len(ctxValues) != 0 {
		values.Set("app_data", c.Env)
	}
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
	var base = "/" + defaultApp.Namespace + "/"
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

// Get a Channel URL for the SDK.
func (c *Context) ChannelURL() string {
	return c.AbsoluteURL("/channel/").String()
}

// Serialize the context back to URL values.
func (c *Context) Values() url.Values {
	values := url.Values{}
	if c.AppID != defaultApp.ID {
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

// JSON representation of Context.
func (c *Context) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
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
	})
}
