// Package browserid provides a way to have a shared identifier for an
// incoming request and allows for it to persist via cookies. Think of
// it like assigning a UUID to each browser/client.
package browserid

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"code.google.com/p/go.net/publicsuffix"
	"github.com/daaku/go.trustforward"
)

// Define a new Cookie via flags. For example, given a name like "browserid",
// you will get these flags and default values along with the crypto.Rand
// reader as Rand:
//
//   browserid.cookie=z
//   browserid.length=16
//   browserid.max-age=87600h
func CookieFlag(name string) *Cookie {
	const cn = "z"
	const tenYears = time.Hour * 24 * 365 * 10
	const length = 16
	c := &Cookie{
		Name:   cn,
		MaxAge: tenYears,
		Length: length,
		Rand:   rand.Reader,
	}

	flag.StringVar(
		&c.Name,
		name+".cookie",
		cn,
		"Name of the cookie to store the ID.",
	)
	flag.DurationVar(
		&c.MaxAge,
		name+".max-age",
		tenYears,
		"Max age of the cookie.",
	)
	flag.UintVar(
		&c.Length,
		name+".len",
		length,
		"Number of random bytes to use for ID.",
	)
	return c
}

type Logger interface {
	Printf(fmt string, args ...interface{})
}

// Cookie provides access to the browserid cookie.
type Cookie struct {
	Name   string        // The name of the cookie.
	MaxAge time.Duration // The life time of the cookie.
	Length uint          // The length in bytes to use for the random id.
	Logger Logger        // Used to log messages about invalid cookie values.
	Rand   io.Reader     // Source of random bytes.
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
		Domain:  c.cookieDomain(trustforward.Host(r)),
	}
	r.AddCookie(cookie)
	http.SetCookie(w, cookie)
	return id
}

func (c *Cookie) genID() string {
	i := make([]byte, c.Length)
	_, err := c.Rand.Read(i)
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
