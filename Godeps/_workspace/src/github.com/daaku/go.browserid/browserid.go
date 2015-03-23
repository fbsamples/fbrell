// Package browserid provides a way to have a shared identifier for an
// incoming request and allows for it to persist via cookies. Think of
// it like assigning a UUID to each browser/client.
package browserid

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.trustforward"
	"github.com/daaku/rell/Godeps/_workspace/src/golang.org/x/net/publicsuffix"
)

var emptyForwarded trustforward.Forwarded

type Logger interface {
	Printf(fmt string, args ...interface{})
}

// Cookie provides access to the browserid cookie.
type Cookie struct {
	Name      string        // The name of the cookie.
	MaxAge    time.Duration // The life time of the cookie.
	Length    uint          // The length in bytes to use for the random id.
	Logger    Logger        // Used to log messages about invalid cookie values.
	Rand      io.Reader     // Source of random bytes.
	Forwarded *trustforward.Forwarded
}

func (c *Cookie) forwarded() *trustforward.Forwarded {
	if c.Forwarded == nil {
		return &emptyForwarded
	}
	return c.Forwarded
}

// Check if a ID has been set.
func (c *Cookie) Has(r *http.Request) bool {
	cookie, err := r.Cookie(c.Name)
	return err == nil && cookie != nil && c.isGood(cookie.Value)
}

// Get the ID, creating one if necessary.
func (c *Cookie) Get(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie(c.Name)
	if err != nil && err != http.ErrNoCookie {
		c.Logger.Printf("Error reading browserid cookie: %s", err)
	}
	if cookie != nil {
		if c.isGood(cookie.Value) {
			return cookie.Value
		}
		c.Logger.Printf("Bad cookie value: %s", cookie.Value)
	}
	id := c.genID()
	cookie = &http.Cookie{
		Name:    c.Name,
		Value:   id,
		Path:    "/",
		Expires: time.Now().Add(c.MaxAge),
		Domain:  c.cookieDomain(c.forwarded().Host(r)),
	}
	r.AddCookie(cookie)
	http.SetCookie(w, cookie)
	return id
}

func (c *Cookie) genID() string {
	i := make([]byte, c.Length)
	r := c.Rand
	if r == nil {
		r = rand.Reader
	}
	_, err := r.Read(i)
	if err != nil {
		panic(fmt.Sprintf("browserid: cookie.Rand.Read failed: %s", err))
	}
	return hex.EncodeToString(i)
}

func (c *Cookie) isGood(value string) bool {
	return uint(len(value)/2) == c.Length
}

// Returns an empty string on failure to skip explicit domain.
func (c *Cookie) cookieDomain(host string) string {
	if strings.Contains(host, ":") {
		h, _, err := net.SplitHostPort(host)
		if err != nil {
			c.Logger.Printf("Error parsing host: %s", host)
			return ""
		}
		host = h
	}
	if host == "localhost" {
		return ""
	}
	if net.ParseIP(host) != nil {
		return ""
	}
	registered, err := publicsuffix.EffectiveTLDPlusOne(host)
	if err != nil {
		c.Logger.Printf("Error extracting base domain: %s", err)
		return ""
	}
	return "." + registered
}
