// Package viewog implements HTTP handlers for /og and /rog* URLs on Rell.
package viewog

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/daaku/go.errcode"
	"github.com/daaku/go.h"
	"github.com/daaku/go.h.js.fb"
	"github.com/daaku/go.static"
	"golang.org/x/net/context"
	"github.com/daaku/rell/og"
	"github.com/daaku/rell/rellenv"
	"github.com/daaku/rell/view"
)

type Handler struct {
	Static       *static.Handler
	ObjectParser *og.Parser
}

// Handles /og/ requests.
func (a *Handler) Values(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	env, err := rellenv.FromContext(ctx)
	if err != nil {
		return err
	}
	values := r.URL.Query()
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) > 4 {
		return errcode.New(http.StatusNotFound, "Invalid URL: %s", r.URL.Path)
	}
	if len(parts) > 2 {
		values.Set("og:type", parts[2])
	}
	if len(parts) > 3 {
		values.Set("og:title", parts[3])
	}
	object, err := a.ObjectParser.FromValues(ctx, env, values)
	if err != nil {
		return err
	}
	h.WriteResponse(w, r, renderObject(ctx, env, a.Static, object))
	return nil
}

// Handles /rog/* requests.
func (a *Handler) Base64(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	env, err := rellenv.FromContext(ctx)
	if err != nil {
		return err
	}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 {
		return errcode.New(http.StatusNotFound, "Invalid URL: %s", r.URL.Path)
	}
	object, err := a.ObjectParser.FromBase64(ctx, env, parts[2])
	if err != nil {
		return err
	}
	h.WriteResponse(w, r, renderObject(ctx, env, a.Static, object))
	return nil
}

// Handles /rog-redirect/ requests.
func (h *Handler) Redirect(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 5 {
		return fmt.Errorf("Invalid URL: %s", r.URL.Path)
	}
	status, err := strconv.Atoi(parts[2])
	if err != nil || (status != 301 && status != 302) {
		return fmt.Errorf("Invalid status: %s", parts[2])
	}
	count, err := strconv.Atoi(parts[3])
	if err != nil {
		return fmt.Errorf("Invalid count: %s", parts[3])
	}
	context, err := rellenv.FromContext(ctx)
	if err != nil {
		return err
	}
	if count == 0 {
		http.Redirect(w, r, context.AbsoluteURL("/rog/"+parts[4]).String(), status)
	} else {
		count--
		url := context.AbsoluteURL(fmt.Sprintf(
			"/rog-redirect/%d/%d/%s", status, count, parts[4]))
		http.Redirect(w, r, url.String(), status)
	}
	return nil
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
func renderObject(ctx context.Context, env *rellenv.Env, s *static.Handler, o *og.Object) h.HTML {
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
					&static.LinkStyle{
						Handler: s,
						HREF:    view.DefaultPageConfig.Style,
					},
					renderMeta(o),
				},
			},
			&h.Body{
				Class: "container",
				Inner: &h.Frag{
					&h.Div{ID: "fb-root"},
					view.DefaultPageConfig.GA,
					&fb.Init{
						URL:   env.SdkURL(),
						AppID: rellenv.FbApp(ctx).ID(),
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
									Class: "btn btn-info pull-right",
									HREF:  o.LintURL(),
									Inner: &h.Frag{
										&h.I{Class: "icon-warning-sign icon-white"},
										h.String(" Debugger"),
									},
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
