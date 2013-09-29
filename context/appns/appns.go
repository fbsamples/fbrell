// Package appns provides a component to get the canvas namespace for a
// Facebook application.
package appns

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/daaku/go.fbapi"
	"github.com/daaku/go.fbapp"
)

type Logger interface {
	Printf(format string, v ...interface{})
}

type Cache interface {
	Get(key string) ([]byte, error)
	Store(key string, val []byte, timeout time.Duration) error
}

type Fetcher struct {
	FbApiClient  *fbapi.Client
	Apps         []fbapp.App
	Logger       Logger
	Cache        Cache
	CacheTimeout time.Duration
}

// Get the App Namespace, fetching it using the Graph API if necessary.
func (c *Fetcher) Get(id uint64) string {
	for _, app := range c.Apps {
		if app.ID() == id {
			return app.Namespace()
		}
	}

	ids := strconv.FormatUint(id, 10)
	ns, _ := c.Cache.Get(ids)
	if ns != nil {
		return string(ns)
	}

	res := struct{ Namespace string }{""}
	req := http.Request{Method: "GET", URL: &url.URL{Path: ids}}
	_, err := c.FbApiClient.Do(&req, &res)
	if err != nil {
		c.Logger.Printf("Ignoring error API call for AppNamespace: %s", err)
		return ""
	}

	c.Cache.Store(ids, []byte(res.Namespace), c.CacheTimeout)
	return res.Namespace
}
