// Package viewexamples provides the various views on Rell examples.
package viewexamples

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/daaku/go.fburl"
	"github.com/daaku/go.h"
	"github.com/daaku/go.h.js.fb"
	"github.com/daaku/go.h.js.loader"
	"github.com/daaku/go.stats"
	"github.com/daaku/go.xsrf"
	"github.com/daaku/rell/context"
	"github.com/daaku/rell/examples"
	"github.com/daaku/rell/js"
	"github.com/daaku/rell/view"
	"net/http"
)

const (
	savedPath = "/saved/"
	paramName = "-xsrf-token-"
)

var (
	envOptions = map[string]string{
		"Production with CDN":    "",
		"Production without CDN": fburl.Production,
		"Beta":                   fburl.Beta,
		"Latest":                 "latest",
		"Dev":                    "dev",
		"Intern":                 "intern",
		"In Your":                "inyour",
	}
	errTokenMismatch = errors.New("Token mismatch.")
)

// Parse the Context and an Example.
func parse(r *http.Request) (*context.Context, *examples.Example, error) {
	context, err := context.FromRequest(r)
	if err != nil {
		return nil, nil, err
	}
	example, err := examples.Load(context.Version, r.URL.Path)
	if err != nil {
		return nil, nil, err
	}
	return context, example, nil
}

// Renders the example content including support for context sensitive
// text substitution.
func content(c *context.Context, e *examples.Example) []byte {
	www := fburl.URL{
		Env: c.Env,
	}
	return bytes.Replace(e.Content, []byte("{{www-server}}"), []byte(www.String()), -1)
}

func List(w http.ResponseWriter, r *http.Request) {
	context, err := context.FromRequest(r)
	if err != nil {
		view.Error(w, r, err)
		return
	}
	stats.Inc("viewed examples listing")
	view.Write(w, r, renderList(context, examples.GetDB(context.Version)))
}

func Saved(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" && r.URL.Path == savedPath {
		c, err := context.FromRequest(r)
		if err != nil {
			view.Error(w, r, err)
			return
		}
		if !xsrf.Validate(r.FormValue(paramName), w, r, savedPath) {
			stats.Inc(savedPath + " xsrf failure")
			view.Error(w, r, errTokenMismatch)
			return
		}
		hash, err := examples.Save([]byte(r.FormValue("code")))
		if err != nil {
			view.Error(w, r, err)
			return
		}
		stats.Inc("saved example")
		http.Redirect(w, r, c.ViewURL(savedPath+hash), 302)
		return
	} else {
		context, example, err := parse(r)
		if err != nil {
			view.Error(w, r, err)
			return
		}
		stats.Inc("viewed saved example")
		view.Write(w, r, &page{
			Writer:  w,
			Request: r,
			Context: context,
			Example: example,
		})
	}
}

func Raw(w http.ResponseWriter, r *http.Request) {
	context, example, err := parse(r)
	if err != nil {
		view.Error(w, r, err)
		return
	}
	if !example.AutoRun {
		view.Error(
			w, r, errors.New("Not allowed to view this example in raw mode."))
		return
	}
	stats.Inc("viewed example in raw mode")
	w.Write(content(context, example))
}

func Simple(w http.ResponseWriter, r *http.Request) {
	context, example, err := parse(r)
	if err != nil {
		view.Error(w, r, err)
		return
	}
	if !example.AutoRun {
		view.Error(
			w, r, errors.New("Not allowed to view this example in simple mode."))
		return
	}
	stats.Inc("viewed example in simple mode")
	view.Write(w, r, &h.Document{
		Inner: &h.Frag{
			&h.Head{
				Inner: &h.Frag{
					&h.Meta{Charset: "utf-8"},
					&h.Title{h.String(example.Title)},
				},
			},
			&h.Body{
				Inner: &h.Frag{
					&loader.HTML{
						Resource: []loader.Resource{
							&fb.Init{
								AppID:      context.AppID,
								ChannelURL: context.ChannelURL(),
								URL:        context.SdkURL(),
							},
						},
					},
					&h.Div{
						ID:    "example",
						Inner: h.UnsafeBytes(content(context, example)),
					},
				},
			},
		},
	})
}

func SdkChannel(w http.ResponseWriter, r *http.Request) {
	const maxAge = 31536000 // 1 year
	context, err := context.FromRequest(r)
	if err != nil {
		view.Error(w, r, err)
		return
	}
	stats.Inc("viewed channel")
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))
	view.Write(w, r, &h.Script{Src: context.SdkURL()})
}

func Example(w http.ResponseWriter, r *http.Request) {
	context, example, err := parse(r)
	if err != nil {
		view.Error(w, r, err)
		return
	}
	stats.Inc("viewed stock example")
	view.Write(w, r, &page{
		Writer:  w,
		Request: r,
		Context: context,
		Example: example,
	})
}

type page struct {
	Writer  http.ResponseWriter
	Request *http.Request
	Context *context.Context
	Example *examples.Example
}

func (p *page) HTML() (h.HTML, error) {
	hiddenInputs := p.Context.Values()
	hiddenInputs.Set(paramName, xsrf.Token(p.Writer, p.Request, savedPath))
	return &view.Page{
		Context: p.Context,
		Title:   p.Example.Title,
		Class:   "main",
		Resource: []loader.Resource{&js.Init{
			Context: p.Context,
			Example: p.Example,
		}},
		Body: &h.Div{
			Class: "container-fluid",
			Inner: &h.Form{
				Action: p.Context.URL(savedPath).String(),
				Method: h.Post,
				Target: "_top",
				Inner: &h.Frag{
					h.HiddenInputs(hiddenInputs),
					&h.Div{
						Class: "row-fluid",
						Inner: &h.Frag{
							&h.Div{
								Class: "span8",
								Inner: &h.Frag{
									&editorTop{Context: p.Context, Example: p.Example},
									&editorArea{Context: p.Context, Example: p.Example},
									&editorBottom{Context: p.Context, Example: p.Example},
									&editorOutput{},
								},
							},
							&h.Div{
								Class: "span4",
								Inner: &h.Frag{
									&logContainer{},
								},
							},
						},
					},
				},
			},
		},
	}, nil
}

type editorTop struct {
	Context *context.Context
	Example *examples.Example
}

func (e *editorTop) HTML() (h.HTML, error) {
	return &h.Div{
		Class: "row-fluid form-inline",
		Inner: &h.Frag{
			&h.Div{
				Class: "span6",
				Inner: &h.Frag{
					&h.A{
						ID:    "rell-login",
						Class: "btn btn-primary",
						Inner: &h.Frag{
							&h.I{Class: "icon-user icon-white"},
							h.String(" Log In"),
						},
					},
					h.String(" "),
					&h.Span{ID: "auth-status-label", Inner: h.String("Status:")},
					h.String(" "),
					&h.Span{ID: "auth-status", Inner: h.String("waiting")},
					h.String(" "),
					&h.Span{Class: "bar", Inner: h.String("|")},
					h.String(" "),
					&h.A{
						ID:    "rell-disconnect",
						Inner: h.String("Disconnect"),
					},
					h.String(" "),
					&h.Span{Class: "bar", Inner: h.String("|")},
					h.String(" "),
					&h.A{
						ID:    "rell-logout",
						Inner: h.String("Logout"),
					},
				},
			},
			&h.Div{
				Class: "span6",
				Inner: &h.Div{
					Class: "pull-right",
					Inner: renderEnvSelector(e.Context, e.Example),
				},
			},
		},
	}, nil
}

type editorArea struct {
	Context *context.Context
	Example *examples.Example
}

func (e *editorArea) HTML() (h.HTML, error) {
	return &h.Div{
		Class: "row-fluid",
		Inner: &h.Textarea{
			ID:    "jscode",
			Name:  "code",
			Inner: h.String(content(e.Context, e.Example)),
		},
	}, nil
}

type editorBottom struct {
	Context *context.Context
	Example *examples.Example
}

func (e *editorBottom) HTML() (h.HTML, error) {
	return &h.Div{
		Class: "row-fluid form-inline",
		Inner: &h.Frag{
			&h.Strong{
				Class: "span4",
				Inner: &h.A{
					HREF:  e.Context.URL("/examples/").String(),
					Inner: h.String("Examples"),
				},
			},
			&h.Div{
				Class: "span8",
				Inner: &h.Div{
					Class: "pull-right",
					Inner: &h.Frag{
						&h.Select{
							ID:   "rell-view-mode",
							Name: "view-mode",
							Inner: &h.Frag{
								&h.Option{
									Inner:    h.String("Website"),
									Selected: e.Context.ViewMode == context.Website,
									Value:    string(context.Website),
									Data: map[string]interface{}{
										"url": e.Context.URL(e.Example.URL).String(),
									},
								},
								&h.Option{
									Inner:    h.String("Canvas"),
									Selected: e.Context.ViewMode == context.Canvas,
									Value:    string(context.Canvas),
									Data: map[string]interface{}{
										"url": e.Context.CanvasURL(e.Example.URL),
									},
								},
								&h.Option{
									Inner:    h.String("Page Tab"),
									Selected: e.Context.ViewMode == context.PageTab,
									Value:    string(context.PageTab),
									Data: map[string]interface{}{
										"url": e.Context.PageTabURL(e.Example.URL),
									},
								},
							},
						},
						h.String(" "),
						&h.Button{
							Class: "btn",
							Type:  "submit",
							Inner: &h.Frag{
								&h.I{Class: "icon-edit"},
								h.String(" Save Code"),
							},
						},
						h.String(" "),
						&h.A{
							ID:    "rell-run-code",
							Class: "btn btn-primary",
							Inner: &h.Frag{
								&h.I{Class: "icon-play icon-white"},
								h.String(" Run Code"),
							},
						},
					},
				},
			},
		},
	}, nil
}

type editorOutput struct{}

func (e *editorOutput) HTML() (h.HTML, error) {
	return &h.Div{Class: "row-fluid", ID: "jsroot"}, nil
}

type logContainer struct{}

func (e *logContainer) HTML() (h.HTML, error) {
	return &h.Div{
		Class: "span4",
		ID:    "log-container",
		Inner: &h.Frag{
			&h.Button{
				ID:    "rell-log-clear",
				Class: "btn",
				Inner: h.String("Clear"),
			},
			&h.Div{ID: "log"},
		},
	}, nil
}

func renderList(context *context.Context, db *examples.DB) *view.Page {
	categories := &h.Frag{}
	for _, category := range db.Category {
		if !category.Hidden {
			categories.Append(renderCategory(context, category))
		}
	}
	return &view.Page{
		Context: context,
		Title:   "Examples",
		Class:   "examples",
		Body: &h.Div{
			Class: "container",
			Inner: &h.Div{
				Class: "row",
				Inner: &h.Div{
					Class: "span12",
					Inner: &h.Frag{
						&h.H1{Inner: h.String("Examples")},
						categories,
					},
				},
			},
		},
	}
}

func renderCategory(c *context.Context, category *examples.Category) h.HTML {
	li := &h.Frag{}
	for _, example := range category.Example {
		li.Append(&h.Li{
			Inner: &h.A{
				HREF:  c.URL(example.URL).String(),
				Inner: h.String(example.Name),
			},
		})
	}
	return &h.Frag{
		&h.H2{Inner: h.String(category.Name)},
		&h.Ul{Inner: li},
	}
}

func renderEnvSelector(c *context.Context, example *examples.Example) h.HTML {
	if !c.IsEmployee {
		return nil
	}
	frag := &h.Frag{}
	foundSelected := false
	selected := false
	for title, value := range envOptions {
		selected = c.Env == value
		if selected {
			foundSelected = true
		}
		ctxCopy := c.Copy()
		ctxCopy.Env = value
		frag.Append(&h.Option{
			Inner:    h.String(title),
			Selected: selected,
			Value:    value,
			Data: map[string]interface{}{
				"url": ctxCopy.ViewURL(example.URL),
			},
		})
	}
	if !foundSelected {
		frag.Append(&h.Option{
			Inner:    h.String(c.Env),
			Selected: true,
			Value:    c.Env,
			Data: map[string]interface{}{
				"url": c.ViewURL(example.URL),
			},
		})
	}
	return &h.Select{
		ID:    "rell-env",
		Name:  "env",
		Inner: frag,
	}
}
