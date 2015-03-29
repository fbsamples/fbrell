// Package appns provides a component to get the canvas namespace for a
// Facebook application.
package appns

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/daaku/rell/Godeps/_workspace/src/github.com/facebookgo/fbapi"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/facebookgo/fbapp"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/golang/groupcache/lru"
)

type cacheKey uint64

type Logger interface {
	Printf(format string, v ...interface{})
}

type Fetcher struct {
	FbApiClient *fbapi.Client
	Apps        []fbapp.App
	Logger      Logger
	Cache       *lru.Cache
}

// Get the App Namespace, fetching it using the Graph API if necessary.
func (c *Fetcher) Get(id uint64) string {
	for _, app := range c.Apps {
		if app.ID() == id {
			return app.Namespace()
		}
	}

	if ns, ok := c.Cache.Get(cacheKey(id)); ok {
		return ns.(string)
	}

	res := struct{ Namespace string }{""}
	req := http.Request{
		Method: "GET",
		URL:    &url.URL{Path: strconv.FormatUint(id, 10)},
	}
	_, err := c.FbApiClient.Do(&req, &res)
	if err != nil {
		c.Logger.Printf("Ignoring error API call for AppNamespace: %s", err)
		return ""
	}

	c.Cache.Add(cacheKey(id), res.Namespace)
	return res.Namespace
}
