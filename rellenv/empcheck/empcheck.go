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

// Package empcheck checks for employees.
package empcheck

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/facebookgo/fbapi"
	"github.com/facebookgo/fbapp"
	"github.com/golang/groupcache/lru"
)

var fields = fbapi.ParamFields("is_employee")

type cacheKey uint64

type user struct {
	IsEmployee bool `json:"is_employee"`
}

type Logger interface {
	Printf(format string, v ...interface{})
}

type Checker struct {
	FbApiClient *fbapi.Client
	App         fbapp.App
	Logger      Logger
	Cache       *lru.Cache
}

// Check if the user is a Facebook Employee. This only available by
// special permission granted to an application by Facebook.
func (c *Checker) Check(id uint64) bool {
	if is, ok := c.Cache.Get(cacheKey(id)); ok {
		return is.(bool)
	}

	values, err := fbapi.ParamValues(c.App, fields)
	if err != nil {
		c.Logger.Printf("Ignoring error in IsEmployee ParamValues: %s", err)
		return false
	}

	var user user
	u := url.URL{
		Path:     strconv.FormatUint(id, 10),
		RawQuery: values.Encode(),
	}
	req := http.Request{Method: "GET", URL: &u}
	_, err = c.FbApiClient.Do(&req, &user)
	if err != nil {
		if apiErr, ok := err.(*fbapi.Error); ok {
			if apiErr.Code == 100 { // common error with test users
				return false
			}
		}
		c.Logger.Printf("Ignoring error in IsEmployee FbApiClient.Do: %s", err)
		return false
	}

	c.Cache.Add(cacheKey(id), user.IsEmployee)
	return user.IsEmployee
}
