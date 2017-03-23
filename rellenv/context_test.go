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

package rellenv_test

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/daaku/go.trustforward"
	"github.com/facebookgo/ensure"
	"github.com/facebookgo/fbapp"
	"github.com/fbsamples/fbrell/rellenv"
)

const (
	defaultFacebookAppID = 42
	defaultAppNS         = "fbrelll"
)

type funcEmpChecker func(uint64) bool

func (f funcEmpChecker) Check(uid uint64) bool {
	return f(uid)
}

type funcAppNSFetcher func(uint64) string

func (f funcAppNSFetcher) Get(id uint64) string {
	return f(id)
}

func defaultParser() *rellenv.Parser {
	return &rellenv.Parser{
		EmpChecker:   funcEmpChecker(func(uint64) bool { return true }),
		AppNSFetcher: funcAppNSFetcher(func(uint64) string { return defaultAppNS }),
		App:          fbapp.New(defaultFacebookAppID, "", ""),
		Forwarded:    &trustforward.Forwarded{},
	}
}

func fromValues(t *testing.T, values url.Values) (*rellenv.Env, context.Context) {
	req, err := http.NewRequest(
		"GET",
		"http://www.fbrell.com/?"+values.Encode(),
		nil)
	if err != nil {
		t.Fatalf("Failed to create request: %s", err)
	}
	env, err := defaultParser().FromRequest(req)
	if err != nil {
		t.Fatalf("Failed to create env: %s", err)
	}
	return env, rellenv.WithEnv(context.Background(), env)
}

func TestDefaultEnv(t *testing.T) {
	t.Parallel()
	env, _ := fromValues(t, url.Values{})
	ensure.Subset(t, env, defaultParser().Default())
}

func TestCustomAppID(t *testing.T) {
	t.Parallel()
	values := url.Values{}
	values.Add("appid", "123")
	_, ctx := fromValues(t, values)
	ensure.DeepEqual(t, rellenv.FbApp(ctx).ID(), uint64(123))
}

func TestCustomStatus(t *testing.T) {
	t.Parallel()
	values := url.Values{}
	values.Add("status", "0")
	env, _ := fromValues(t, values)
	if env.Status {
		t.Fatal("Was expecting status to be false.")
	}
}

func TestComplex(t *testing.T) {
	t.Parallel()
	values := url.Values{}
	values.Add("status", "1")
	values.Add("server", "beta")
	values.Add("locale", "en_PI")
	expected := &rellenv.Env{
		Status: true,
		Env:    "beta",
	}
	env, _ := fromValues(t, values)
	ensure.Subset(t, env, expected)
	ensure.StringContains(t, env.SdkURL(), "en_PI")
}

func TestPageTabURLBeta(t *testing.T) {
	t.Parallel()
	env, _ := fromValues(t, url.Values{"server": []string{"beta"}})
	pageTabURL := env.PageTabURL("/")
	ensure.StringContains(t, pageTabURL,
		"http://www.beta.facebook.com/pages/Rell-Page-for-Tabs/141929622497380")
}

func TestPageTabURL(t *testing.T) {
	t.Parallel()
	env, _ := fromValues(t, url.Values{})
	pageTabURL := env.PageTabURL("/")
	ensure.StringContains(t, pageTabURL,
		"http://www.facebook.com/pages/Rell-Page-for-Tabs/141929622497380")
	ensure.StringContains(t, pageTabURL, fmt.Sprintf("app_%d", defaultFacebookAppID))
	ensure.StringContains(t, pageTabURL, "app_data=Lw%3D%3D")
}

func TestCanvasURLBeta(t *testing.T) {
	t.Parallel()
	env, _ := fromValues(t, url.Values{"server": []string{"beta"}})
	canvasURL := env.CanvasURL("/")
	ensure.StringContains(t, canvasURL,
		fmt.Sprintf("https://apps.beta.facebook.com/%s/?server=beta", defaultAppNS))
}

func TestCanvasURL(t *testing.T) {
	t.Parallel()
	env, _ := fromValues(t, url.Values{})
	canvasURL := env.CanvasURL("/")
	ensure.StringContains(t, canvasURL,
		fmt.Sprintf("https://apps.facebook.com/%s/", defaultAppNS))
}
