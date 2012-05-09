// Package viewog implements HTTP handlers for /og and /rog* URLs on Rell.
package viewog

import (
	"fmt"
	"github.com/nshah/go.h"
	"github.com/nshah/go.h.js.fb"
	"github.com/nshah/go.h.js.loader"
	"github.com/nshah/rell/context"
	"github.com/nshah/rell/og"
	"github.com/nshah/rell/view"
	"net/http"
	"strings"
)

// Handles /og/ requests.
func Values(w http.ResponseWriter, r *http.Request) {
	context, err := context.FromRequest(r)
	if err != nil {
		view.Error(w, err)
		return
	}
	values := r.URL.Query()
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) > 4 {
		view.Error(w, fmt.Errorf("Invalid URL: %s", r.URL.Path))
		return
	}
	if len(parts) > 2 {
		values.Set("og:type", parts[2])
	}
	if len(parts) > 3 {
		values.Set("og:title", parts[3])
	}
	object, err := og.NewFromValues(context, values)
	if err != nil {
		view.Error(w, err)
		return
	}
	h.Write(w, renderObject(context, object))
}

// Handles /rog/* requests.
func Base64(w http.ResponseWriter, r *http.Request) {
	context, err := context.FromRequest(r)
	if err != nil {
		view.Error(w, err)
		return
	}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 {
		view.Error(w, fmt.Errorf("Invalid URL: %s", r.URL.Path))
		return
	}
	object, err := og.NewFromBase64(context, parts[2])
	if err != nil {
		view.Error(w, err)
		return
	}
	h.Write(w, renderObject(context, object))
}

// Handles /rog-redirect/ requests.
func Redirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "", 302)
}

// Renders <meta> tags for object.
func renderMeta(o *og.Object) h.HTML {
	frag := &h.Frag{}
	for _, pair := range o.Pairs {
		frag.Append(&h.Meta{
			Property: pair.Key,
			Content:  pair.Value,
		})
	}
	return frag
}

// Auto linkify values that start with "http".
func renderValue(val string) h.HTML {
	txt := h.String(val)
	if strings.HasPrefix(val, "http") {
		return &h.A{HREF: val, Inner: txt}
	}
	return txt
}

// Renders a <table> with the meta data for the object.
func renderMetaTable(o *og.Object) h.HTML {
	frag := &h.Frag{}
	for _, pair := range o.Pairs {
		frag.Append(&h.Tr{
			Inner: &h.Frag{
				&h.Th{Inner: h.String(pair.Key)},
				&h.Td{Inner: renderValue(pair.Value)},
			},
		})
	}

	return &h.Table{
		Class: "info",
		Inner: &h.Frag{
			&h.Thead{
				Inner: &h.Tr{
					Inner: &h.Frag{
						&h.Th{Inner: h.String("Property")},
						&h.Th{Inner: h.String("Content")},
					},
				},
			},
			&h.Tbody{Inner: frag},
		},
	}
}

// Render a document for the Object.
func renderObject(context *context.Context, o *og.Object) h.HTML {
	return &h.Document{
		Inner: &h.Frag{
			&h.Head{
				Inner: &h.Frag{
					&h.Meta{Charset: "utf-8"},
					&h.Title{h.String(o.Title())},
					view.DefaultStyle,
					renderMeta(o),
				},
			},
			&h.Body{
				Class: "narrow",
				Inner: &h.Frag{
					&h.Div{ID: "fb-root"},
					&loader.HTML{
						Resource: []loader.Resource{
							view.DefaultGA,
							&fb.Init{
								URL:        context.SdkURL(),
								AppID:      context.AppID,
								ChannelURL: context.ChannelURL(),
							},
						},
					},
					&h.Div{
						Class: "bd",
						Inner: &h.Frag{
							&h.H1{
								Inner: &h.A{
									HREF:  o.URL(),
									Inner: h.String(o.Title()),
								},
							},
							renderMetaTable(o),
							&h.A{
								HREF: o.ImageURL(),
								Inner: &h.Img{
									Src: o.ImageURL(),
								},
							},
							&h.A{
								Class: "lint-this",
								HREF:  o.LintURL(),
								Inner: h.String("Lint this."),
							},
							&h.Iframe{
								Class:             "like",
								Src:               o.LikeURL(),
								Scrolling:         false,
								FrameBorder:       0,
								AllowTransparency: true,
							},
						},
					},
				},
			},
		},
	}
}
