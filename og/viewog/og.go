// Package viewog implements HTTP handlers for /og and /rog* URLs on Rell.
package viewog

import (
	"fmt"
	"github.com/daaku/go.errcode"
	"github.com/daaku/go.h"
	"github.com/daaku/go.h.js.fb"
	"github.com/daaku/go.h.js.loader"
	"github.com/daaku/go.stats"
	"github.com/daaku/rell/context"
	"github.com/daaku/rell/og"
	"github.com/daaku/rell/view"
	"net/http"
	"strconv"
	"strings"
)

// Handles /og/ requests.
func Values(w http.ResponseWriter, r *http.Request) {
	context, err := context.FromRequest(r)
	if err != nil {
		view.Error(w, r, err)
		return
	}
	values := r.URL.Query()
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) > 4 {
		view.Error(w, r, errcode.New(
			http.StatusNotFound, "Invalid URL: %s", r.URL.Path))
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
		view.Error(w, r, err)
		return
	}
	stats.Inc("viewed og")
	view.Write(w, r, renderObject(context, object))
}

// Handles /rog/* requests.
func Base64(w http.ResponseWriter, r *http.Request) {
	context, err := context.FromRequest(r)
	if err != nil {
		view.Error(w, r, err)
		return
	}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 {
		view.Error(w, r, errcode.New(
			http.StatusNotFound, "Invalid URL: %s", r.URL.Path))
		return
	}
	object, err := og.NewFromBase64(context, parts[2])
	if err != nil {
		view.Error(w, r, err)
		return
	}
	stats.Inc("viewed rog")
	view.Write(w, r, renderObject(context, object))
}

// Handles /rog-redirect/ requests.
func Redirect(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 5 {
		view.Error(w, r, fmt.Errorf("Invalid URL: %s", r.URL.Path))
		return
	}
	status, err := strconv.Atoi(parts[2])
	if err != nil || (status != 301 && status != 302) {
		view.Error(w, r, fmt.Errorf("Invalid status: %s", parts[2]))
		return
	}
	count, err := strconv.Atoi(parts[3])
	if err != nil {
		view.Error(w, r, fmt.Errorf("Invalid count: %s", parts[3]))
		return
	}
	context, err := context.FromRequest(r)
	if err != nil {
		view.Error(w, r, err)
		return
	}
	stats.Inc("rog-redirect request")
	if count == 0 {
		http.Redirect(
			w, r, context.AbsoluteURL("/rog/"+parts[4]).String(), status)
	} else {
		count--
		url := context.AbsoluteURL(fmt.Sprintf(
			"/rog-redirect/%d/%d/%s", status, count, parts[4]))
		http.Redirect(w, r, url.String(), status)
	}
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
		Class: "table table-bordered table-striped og-info",
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
	var title, header h.HTML
	if o.Title() != "" {
		title = &h.Title{h.String(o.Title())}
		header = &h.H1{
			Inner: &h.A{
				HREF:  o.URL(),
				Inner: h.String(o.Title()),
			},
		}
	}
	return &h.Document{
		Inner: &h.Frag{
			&h.Head{
				Inner: &h.Frag{
					&h.Meta{Charset: "utf-8"},
					title,
					view.DefaultStyle,
					renderMeta(o),
				},
			},
			&h.Body{
				Class: "container",
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
						Class: "row",
						Inner: &h.Frag{
							&h.Div{
								Class: "span8",
								Inner: header,
							},
							&h.Div{
								Class: "span4",
								Inner: &h.A{
									Class: "lint-this btn btn-primary pull-right",
									HREF:  o.LintURL(),
									Inner: h.String("Lint this."),
								},
							},
						},
					},
					&h.Div{
						Class: "row",
						Inner: &h.Frag{
							&h.Div{
								Class: "span6",
								Inner: &h.Frag{
									renderMetaTable(o),
									&h.Iframe{
										Class: "like",
										Src:   o.LikeURL(),
									},
								},
							},
							&h.Div{
								Class: "span6",
								Inner: &h.A{
									HREF: o.ImageURL(),
									Inner: &h.Img{
										Src: o.ImageURL(),
										Alt: o.Title(),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
