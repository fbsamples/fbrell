// Package trustforward provides wrappers that trust X-Forwarded-*
// headers when looking up certain values. Using this package to access
// the values allows you to control at runtime (via a flag) whether the
// values will be trusted.
package trustforward

import "net/http"

// Forwarded enables or disables X-Forwarded or CloudFlare headers.
type Forwarded struct {
	X          bool
	CloudFlare bool
}

// Get the Host.
func (f *Forwarded) Host(r *http.Request) string {
	if f.X {
		if fwdHost := r.Header.Get("x-forwarded-host"); fwdHost != "" {
			return fwdHost
		}
	}
	return r.Host
}

// Get the Scheme.
func (f *Forwarded) Scheme(r *http.Request) string {
	if f.CloudFlare {
		const cfHttps = `{"scheme":"https"}`
		if cfVisitor := r.Header.Get("Cf-Visitor"); cfVisitor == cfHttps {
			return "https"
		}
	}
	if f.X {
		if fwdScheme := r.Header.Get("x-forwarded-proto"); fwdScheme != "" {
			return fwdScheme
		}
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

// Get the Remote Address.
func (f *Forwarded) Remote(r *http.Request) string {
	if f.CloudFlare {
		if cfConnectingIp := r.Header.Get("Cf-Connecting-Ip"); cfConnectingIp != "" {
			return cfConnectingIp
		}
	}
	if f.X {
		if fwdRemote := r.Header.Get("x-forwarded-for"); fwdRemote != "" {
			return fwdRemote
		}
	}
	return r.RemoteAddr
}
