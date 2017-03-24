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

// Package view implements the generic Rell view logic including the
// standard base page, error page and so on.
package view

import (
	"context"

	"github.com/daaku/go.h"
	"github.com/daaku/go.h.js.ga"
	"github.com/daaku/go.static"
)

type PageConfig struct {
	GA     *ga.Track
	Style  []string
	Script []string
}

var DefaultPageConfig = &PageConfig{
	GA: &ga.Track{Account: "UA-15507059-1"},
	Style: []string{
		"css/rell.css",
	},
	Script: []string{
		"js/log.js",
		"js/rell.js",
	},
}

// A minimal standard page with no visible body.
type Page struct {
	Config *PageConfig
	Class  string
	Head   h.HTML
	Body   h.HTML
	Title  string
}

func (p *Page) config() *PageConfig {
	if p.Config == nil {
		return DefaultPageConfig
	}
	return p.Config
}

func (p *Page) HTML(ctx context.Context) (h.HTML, error) {
	return &h.Document{
		XMLNS: h.XMLNS{"fb": "http://ogp.me/ns/fb#"},
		Inner: h.Frag{
			&h.Head{
				Inner: h.Frag{
					&h.Meta{Charset: "utf-8"},
					&h.Meta{Name: "viewport", Content: "width=device-width,initial-scale=1.0"},
					&h.Title{
						h.String(p.Title),
						h.Unsafe(" &mdash; Facebook Read Eval Log Loop"),
					},
					&h.LinkStyle{
						HREF: "https://maxcdn.bootstrapcdn.com/twitter-bootstrap/2.2.0/css/bootstrap-combined.min.css",
					},
					&static.LinkStyle{
						HREF: p.config().Style,
					},
					p.Head,
				},
			},
			&h.Body{
				Class: p.Class,
				Inner: h.Frag{
					p.Body,
					&h.Div{ID: "fb-root"},
					&h.Div{ID: "FB_HiddenContainer"},
					&h.Script{
						Src: "https://maxcdn.bootstrapcdn.com/twitter-bootstrap/2.2.0/js/bootstrap.min.js",
					},
					&h.Script{
						Src: "https://ajax.googleapis.com/ajax/libs/jquery/1.8.2/jquery.min.js",
					},
					&static.Script{
						Src:   p.config().Script,
					},
					p.config().GA,
				},
			},
		},
	}, nil
}
