// Package fburl provides a convenient way to generate facebook.com
// URLs. It allows for generating URLs for various subdomains, and is
// also aware of "environments" which allows switching between
// production and development environments.
//
// Additionally it allows for the default environment be configured
// via a command line flag allowing switching the environment at
// runtime for the entire application.
package fburl

import (
	"flag"
	"net/url"
	"strings"
)

// Commonly used environments.
const (
	Beta       = "beta"
	Production = "www"
)

// Commonly used subdomains.
const (
	DApi        = "api"
	DApps       = "apps"
	DDevelopers = "developers"
	DGraph      = "graph"
	DMobile     = "m"
	DWww        = "www"
)

var defaultEnv = flag.String(
	"fburl.env",
	Production,
	"Facebook environment to generate URLs for.",
)

// Used to construct facebook.com URLs.
type URL struct {
	Scheme    string // Defaults to http.
	Path      string // Defaults to /.
	SubDomain string // Defaults to www.
	Values    url.Values
	Env       string // Default via flag, or else production.
}

// Returns the first non empty string or finally the empty string.
func defaultString(choices ...string) string {
	for _, c := range choices {
		if c != "" {
			return c
		}
	}
	return ""
}

// Returns the default Facebook environment.
func DefaultEnv() string {
	return *defaultEnv
}

// Make a hostname for the given subdomain and environment.
func Hostname(subDomain string, env string) string {
	if !strings.HasPrefix(env, DWww) {
		env = DWww + "." + env
	}
	return strings.Replace(env+".facebook.com", DWww, subDomain, 1)
}

// Generate a url.URL object for the Facebook URL.
func (u *URL) URL() *url.URL {
	env := defaultString(u.Env, DefaultEnv())
	subDomain := u.SubDomain
	if subDomain == "our" && env == Production {
		subDomain = "our.intern"
	}
	return &url.URL{
		Scheme:   defaultString(u.Scheme, "http"),
		Path:     defaultString(u.Path, "/"),
		RawQuery: u.Values.Encode(),
		Host:     Hostname(defaultString(subDomain, DWww), env),
	}
}

// Generate a string for the URL.
func (u *URL) String() string {
	return u.URL().String()
}
