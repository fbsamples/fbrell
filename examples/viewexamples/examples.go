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
	"github.com/daaku/go.h.ui"
	"github.com/daaku/go.stats"
	"github.com/daaku/go.xsrf"
	"github.com/daaku/rell/context"
	"github.com/daaku/rell/examples"
	"github.com/daaku/rell/js"
	"github.com/daaku/rell/view"
	"net/http"
	"net/url"
)

const (
	savedPath = "/saved/"
	paramName = "-xsrf-token-"
)

var (
	envOptions = map[string]string{
		"":               "Production with CDN",
		fburl.Production: "Production without CDN",
		fburl.Beta:       "Beta",
		"latest":         "Latest",
		"dev":            "Dev",
		"intern":         "Intern",
		"inyour":         "In Your",
	}
	viewModeOptions = map[string]string{
		context.Website: "Website",
		context.PageTab: "Page Tab",
		context.Canvas:  "Canvas",
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
	view.Write(w, r, &examplesList{
		Context: context,
		DB:      examples.GetDB(context.Version),
	})
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
				Action: savedPath,
				Method: h.Post,
				Target: "_top",
				Inner: &h.Frag{
					h.HiddenInputs(url.Values{
						paramName: []string{xsrf.Token(p.Writer, p.Request, savedPath)},
					}),
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
									&contextEditor{Context: p.Context, Example: p.Example},
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
					Inner: &envSelector{
						Context: e.Context,
						Example: e.Example,
					},
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

type viewModeDropdown struct {
	Context *context.Context
	Example *examples.Example
}

func (d *viewModeDropdown) HTML() (h.HTML, error) {
	return &h.Div{
		Class: "btn-group",
		Inner: &h.Frag{
			&h.Div{
				Style: "display:none",
				Inner: &h.Input{
					Type:  "hidden",
					ID:    "rell-view-mode",
					Name:  "view-mode",
					Value: d.Context.ViewMode,
				},
			},
			&h.Button{
				Class: "btn",
				Inner: h.String(viewModeOptions[d.Context.ViewMode]),
			},
			&h.Button{
				Class: "btn dropdown-toggle",
				Data: map[string]interface{}{
					"toggle": "dropdown",
				},
				Inner: &h.Span{
					Class: "caret",
				},
			},
			&h.Ul{
				Class: "dropdown-menu",
				Inner: &h.Frag{
					&h.Li{
						Inner: &h.A{
							Inner:  h.String(viewModeOptions[context.Website]),
							Target: "_top",
							HREF:   d.Context.URL(d.Example.URL).String(),
						},
					},
					&h.Li{
						Inner: &h.A{
							Inner:  h.String(viewModeOptions[context.Canvas]),
							Target: "_top",
							HREF:   d.Context.CanvasURL(d.Example.URL),
						},
					},
					&h.Li{
						Inner: &h.A{
							Inner:  h.String(viewModeOptions[context.PageTab]),
							Target: "_top",
							HREF:   d.Context.PageTabURL(d.Example.URL),
						},
					},
				},
			},
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
					Class: "btn-toolbar pull-right",
					Inner: &h.Frag{
						&viewModeDropdown{
							Context: e.Context,
							Example: e.Example,
						},
						h.String(" "),
						&h.Div{
							Class: "btn-group",
							Inner: &h.Button{
								Class: "btn",
								Type:  "submit",
								Inner: &h.Frag{
									&h.I{Class: "icon-edit"},
									h.String(" Save Code"),
								},
							},
						},
						h.String(" "),
						&h.Div{
							Class: "btn-group",
							Inner: &h.A{
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
		ID: "log-container",
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

type contextEditor struct {
	Context *context.Context
	Example *examples.Example
}

func (e *contextEditor) HTML() (h.HTML, error) {
	if !e.Context.IsEmployee {
		return h.HiddenInputs(e.Context.Values()), nil
	}
	return &h.Div{
		Class: "well form-horizontal",
		Inner: &h.Frag{
			&ui.TextInput{
				Label:      h.String("Application ID"),
				Name:       "appid",
				Value:      e.Context.AppID,
				InputClass: "input-medium",
			},
			&ui.ToggleGroup{
				Inner: &h.Frag{
					&ui.ToggleItem{
						Name:        "init",
						Checked:     e.Context.Init,
						Description: h.String("Automatically initialize SDK."),
					},
					&ui.ToggleItem{
						Name:        "status",
						Checked:     e.Context.Status,
						Description: h.String("Automatically trigger status ping."),
					},
					&ui.ToggleItem{
						Name:        "channel",
						Checked:     e.Context.UseChannel,
						Description: h.String("Specify explicit XD channel."),
					},
					&ui.ToggleItem{
						Name:        "frictionlessRequests",
						Checked:     e.Context.FrictionlessRequests,
						Description: h.String("Enable frictionless requests."),
					},
				},
			},
			&h.Div{
				Class: "form-actions",
				Inner: &h.Frag{
					&h.Button{
						Type:  "submit",
						Class: "btn btn-primary",
						Inner: &h.Frag{
							&h.I{Class: "icon-refresh icon-white"},
							h.String(" Update"),
						},
					},
				},
			},
		},
	}, nil
}

type examplesList struct {
	Context *context.Context
	DB      *examples.DB
}

func (l *examplesList) HTML() (h.HTML, error) {
	categories := &h.Frag{}
	for _, category := range l.DB.Category {
		if !category.Hidden {
			categories.Append(&exampleCategory{
				Context:  l.Context,
				Category: category,
			})
		}
	}
	return &view.Page{
		Context: l.Context,
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
	}, nil
}

type exampleCategory struct {
	Context  *context.Context
	Category *examples.Category
}

func (c *exampleCategory) HTML() (h.HTML, error) {
	li := &h.Frag{}
	for _, example := range c.Category.Example {
		li.Append(&h.Li{
			Inner: &h.A{
				HREF:  c.Context.URL(example.URL).String(),
				Inner: h.String(example.Name),
			},
		})
	}
	return &h.Frag{
		&h.H2{Inner: h.String(c.Category.Name)},
		&h.Ul{Inner: li},
	}, nil
}

type envSelector struct {
	Context *context.Context
	Example *examples.Example
}

func (e *envSelector) HTML() (h.HTML, error) {
	if !e.Context.IsEmployee {
		return nil, nil
	}
	frag := &h.Frag{
		h.HiddenInputs(url.Values{
			"server": []string{e.Context.Env},
		}),
	}
	for value, title := range envOptions {
		if e.Context.Env == value {
			continue
		}
		ctxCopy := e.Context.Copy()
		ctxCopy.Env = value
		frag.Append(&h.Li{
			Inner: &h.A{
				Inner:  h.String(title),
				Target: "_top",
				HREF:   ctxCopy.ViewURL(e.Example.URL),
			},
		})
	}

	title := envOptions[e.Context.Env]
	if title == "" {
		title = e.Context.Env
	}
	return &h.Div{
		Class: "btn-group",
		Inner: &h.Frag{
			&h.Button{
				Class: "btn",
				Inner: h.String(title),
			},
			&h.Button{
				Class: "btn dropdown-toggle",
				Data: map[string]interface{}{
					"toggle": "dropdown",
				},
				Inner: &h.Span{
					Class: "caret",
				},
			},
			&h.Ul{
				Class: "dropdown-menu",
				Inner: frag,
			},
		},
	}, nil
}
