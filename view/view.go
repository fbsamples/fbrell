// Package view implements the generic Rell view logic including the
// standard base page, error page and so on.
package view

import (
	"github.com/nshah/go.h"
	"github.com/nshah/go.h.js.ga"
	"github.com/nshah/go.h.js.loader"
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
var DefaultStyle = h.Compile(&h.LinkStyle{HREF: "/public/rell.css"})

// The default Google Analytics setup.
var DefaultGA = &ga.Track{ID: "UA-15507059-1"}

// A minimal standard page with no visible body.
type Page struct {
	Class    string
	Head     h.HTML
	Body     h.HTML
	Title    string
	Resource []loader.Resource
}

func (p *Page) HTML() (h.HTML, error) {
	return &h.Document{
		XMLNS: h.XMLNS{"fb": "http://ogp.me/ns/fb#"},
		Inner: &h.Frag{
			&h.Head{
				Inner: &h.Frag{
					&h.Meta{Charset: "utf-8"},
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
				},
			},
		},
	}, nil
}

// Send a 500 error response.
func Error(w http.ResponseWriter, err error) {
	w.WriteHeader(500)
	page := &Page{
		Body: h.String(err.Error()),
	}
	h.Write(w, page)
}

// Send a 404 response.
func NotFound(w http.ResponseWriter, err error) {
	w.WriteHeader(404)
	page := &Page{
		Body: h.String(err.Error()),
	}
	h.Write(w, page)
}
