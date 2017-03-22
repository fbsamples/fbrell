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

package og

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/daaku/go.static"
	"github.com/facebookgo/fbapp"
	"github.com/fbsamples/fbrell/rellenv"
)

var defaultContext = (&rellenv.Parser{
	App: fbapp.New(0, "", ""),
}).Default()

func defaultParser() *Parser {
	return &Parser{
		Static: &static.Handler{
			Path: "/static/",
			Box:  static.FileSystemBox(http.Dir("../public")),
		},
	}
}

// Order insensitive pairs matching. This isn't fully accurate as OG
// is order sensitive. But since query parameters are not, we use this
// to ignore order.
func assertSubset(t *testing.T, expected, actual *Object) {
Outer:
	for _, pair := range expected.Pairs {
		for _, value := range actual.GetAll(pair.Key) {
			if pair.Value == value {
				continue Outer
			}
		}
		t.Fatalf(
			"Did not find expected pair %q = %q in\n%+v",
			pair.Key,
			pair.Value,
			actual)
	}
}

func TestParseBase64(t *testing.T) {
	t.Parallel()
	const song1 = "W1sib2c6dGl0bGUiLCJzb25nMSJdLFsib2c6dHlwZSIsInNvbmciXV0"
	expected := &Object{Pairs: []Pair{
		{"og:title", "song1"},
		{"og:type", "song"},
		{"og:url", "http://www.fbrell.com/rog/" + song1},
		{"og:image", "http://www.fbrell.com/static/W1siL2ltYWdlcy90YXhpX3JvdGlhXzI4MDYzMzkxMjUuanBnIiwiMTdkMTlmNDUiXV0.jpg"},
		{"og:description", stockDescriptions[0]},
	}}

	object, err := defaultParser().FromBase64(context.Background(), defaultContext, song1)
	if err != nil {
		t.Fatal(err)
	}
	assertSubset(t, expected, object)
}

func TestParseValues(t *testing.T) {
	t.Parallel()
	const ogType = "article"
	const ogTitle = "foo"
	values := url.Values{}
	values.Set("og:type", ogType)
	values.Set("og:title", ogTitle)
	expected := &Object{Pairs: []Pair{
		{"og:type", ogType},
		{"og:title", ogTitle},
		{"og:url", "http://www.fbrell.com/og/" + ogType + "/" + ogTitle},
		{"og:image", "http://www.fbrell.com/static/W1siL2ltYWdlcy90YXhpX3JvdGlhXzI4MDYzMzkxMjUuanBnIiwiMTdkMTlmNDUiXV0.jpg"},
		{"og:description", stockDescriptions[6]},
	}}

	object, err := defaultParser().FromValues(context.Background(), defaultContext, values)
	if err != nil {
		t.Fatal(err)
	}
	assertSubset(t, expected, object)
}
