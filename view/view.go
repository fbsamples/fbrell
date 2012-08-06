// Package view implements the generic Rell view logic including the
// standard base page, error page and so on.
package view

import (
	"github.com/daaku/go.h"
	"github.com/daaku/go.h.js.ga"
	"github.com/daaku/go.h.js.loader"
	"github.com/daaku/go.static"
	"github.com/daaku/rell/context"
	"log"
	"net/http"
)

// The default metadata.
var DefaultMeta = h.Compile(&h.Frag{
	&h.Meta{
		Name:    "description",
		Content: "Examples for the Facebook Platform.",
	},
	&h.Meta{
		Name: "keywords",
		Content: "facebook, connect, facebook connect, javascript, " +
			"examples, javascript sdk, javascript library, library, howto, " +
			"tutorial, api, facebook api, authorization, xfbml, fbml, xfbml " +
			"tags, fbml tags, facebook platform, facebook rest api, graph, " +
			"facebook graph api, facebook graph api examples, facebook old" +
			"rest api, facebook sdk",
	},
	&h.Meta{Property: "fb:app_id", Content: "184484190795"},
	&h.Meta{Property: "og:type", Content: "website"},
	&h.Meta{Property: "og:url", Content: "http://www.fbrell.com/"},
	&h.Meta{
		Property: "og:image",
		Content:  "https://www.fbrell.com/logo.jpg",
	},
})

// The default stylesheet.
var DefaultStyle = &h.Frag{
	&static.LinkStyle{HREF: "css/bootstrap.min.css"},
	&static.LinkStyle{HREF: "css/rell.css"},
}

// The default Google Analytics setup.
var DefaultGA = &ga.Track{ID: "UA-15507059-1"}

// Bootstrap Scripts.
var BootstrapScripts = &h.Frag{
	&static.Script{Src: "js/jquery-1.7.2.min.js"},
	&static.Script{Src: "js/bootstrap.min.js"},
}

// A minimal standard page with no visible body.
type Page struct {
	Context  *context.Context
	Class    string
	Head     h.HTML
	Body     h.HTML
	Title    string
	Resource []loader.Resource
}

func (p *Page) viewport() h.HTML {
	if p.Context != nil {
		viewportContent := p.Context.Viewport()
		if viewportContent != "" {
			return &h.Meta{Name: "viewport", Content: viewportContent}
		}
	}
	return nil
}

func (p *Page) HTML() (h.HTML, error) {
	return &h.Document{
		XMLNS: h.XMLNS{"fb": "http://ogp.me/ns/fb#"},
		Inner: &h.Frag{
			&h.Head{
				Inner: &h.Frag{
					&h.Meta{Charset: "utf-8"},
					p.viewport(),
					&h.Title{
						h.String(p.Title),
						h.Unsafe(" &mdash; Facebook Read Eval Log Loop"),
					},
					DefaultStyle,
					p.Head,
				},
			},
			&h.Body{
				Class: p.Class,
				Inner: &h.Frag{
					p.Body,
					&h.Div{ID: "fb-root"},
					&h.Div{ID: "FB_HiddenContainer"},
					&loader.HTML{
						Resource: append([]loader.Resource{DefaultGA}, p.Resource...),
					},
					BootstrapScripts,
				},
			},
		},
	}, nil
}

// Writes a HTML response and writes errors on failure.
func Write(w http.ResponseWriter, r *http.Request, html h.HTML) {
	if r.Method != "HEAD" {
		_, err := h.Write(w, html)
		if err != nil {
			log.Printf("Error writing HTML for URL: %s: %s", r.URL, err)
			h.Write(w, h.String("FATAL ERROR"))
		}
	}
}
