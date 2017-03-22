/**
 * Copyright (c) 2014-present, Facebook, Inc. All rights reserved.
 *
 * You are hereby granted a non-exclusive, worldwide, royalty-free license to use,
 * copy, modify, and distribute this software in source code or binary form for use
 * in connection with the web services and APIs provided by Facebook.
 *
 * As with any software that integrates with the Facebook platform, your use of
 * this software is subject to the Facebook Developer Principles and Policies
 * [http://developers.facebook.com/policy/]. This copyright notice shall be
 * included in all copies or substantial portions of the software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

// Package appns provides a component to get the canvas namespace for a
// Facebook application.
package appns

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/facebookgo/fbapi"
	"github.com/facebookgo/fbapp"
	"github.com/golang/groupcache/lru"
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
