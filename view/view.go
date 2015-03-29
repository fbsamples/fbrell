// Package view implements the generic Rell view logic including the
// standard base page, error page and so on.
package view

import (
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.h"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.h.js.ga"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.static"
	"github.com/daaku/rell/rellenv"
)

type PageConfig struct {
	GA     *ga.Track
	Style  []string
	Script []string
}

var DefaultPageConfig = &PageConfig{
	GA: &ga.Track{ID: "UA-15507059-1"},
	Style: []string{
		"css/bootstrap.min.css",
		"css/bootstrap-responsive.min.css",
		"css/rell.css",
	},
	Script: []string{
		"js/jquery-1.8.2.min.js",
		"js/bootstrap.min.js",
		"js/log.js",
		"js/rell.js",
	},
}

// A minimal standard page with no visible body.
type Page struct {
	Config  *PageConfig
	Context *rellenv.Context
	Static  *static.Handler
	Class   string
	Head    h.HTML
	Body    h.HTML
	Title   string
}

func (p *Page) config() *PageConfig {
	if p.Config == nil {
		return DefaultPageConfig
	}
	return p.Config
}

func (p *Page) HTML() (h.HTML, error) {
	return &h.Document{
		XMLNS: h.XMLNS{"fb": "http://ogp.me/ns/fb#"},
		Inner: &h.Frag{
			&h.Head{
				Inner: &h.Frag{
					&h.Meta{Charset: "utf-8"},
					&h.Meta{Name: "viewport", Content: "width=device-width,initial-scale=1.0"},
					&h.Title{
						h.String(p.Title),
						h.Unsafe(" &mdash; Facebook Read Eval Log Loop"),
					},
					&static.LinkStyle{
						Handler: p.Static,
						HREF:    p.config().Style,
					},
					p.Head,
				},
			},
			&h.Body{
				Class: p.Class,
				Inner: &h.Frag{
					p.Body,
					&h.Div{ID: "fb-root"},
					&h.Div{ID: "FB_HiddenContainer"},
					&static.Script{
						Handler: p.Static,
						Src:     p.config().Script,
						Async:   true,
					},
					p.config().GA,
				},
			},
		},
	}, nil
}
