// Package viewexamples provides the various views on Rell examples.
package viewexamples

import (
	"bytes"
	"crypto/rand"
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
	"io"
	"log"
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

func makeID(prefix string) string {
	b := make([]byte, 8)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		log.Fatal(err)
	}
	if prefix == "" {
		return fmt.Sprintf("%x", b)
	}
	return fmt.Sprintf("%s_%x", prefix, b)
}

type textInput struct {
	Type  string
	Label h.HTML
	Name  string
	Value interface{}
}

func (i *textInput) HTML() (h.HTML, error) {
	t := i.Type
	if t == "" {
		t = "text"
	}
	id := makeID(i.Name)
	return &h.Div{
		Class: "control-group",
		Inner: &h.Frag{
			&h.Label{
				Class: "control-label",
				For:   id,
				Inner: i.Label,
			},
			&h.Div{
				Class: "controls",
				Inner: &h.Frag{
					&h.Input{
						Type:  t,
						ID:    id,
						Name:  i.Name,
						Value: fmt.Sprint(i.Value),
					},
				},
			},
		},
	}, nil
}

type checkboxInput struct {
	Label       h.HTML
	Name        string
	Checked     bool
	Description h.HTML
}

func (i *checkboxInput) HTML() (h.HTML, error) {
	id := makeID(i.Name)
	return &h.Div{
		Class: "control-group",
		Inner: &h.Frag{
			&h.Label{
				Class: "control-label",
				For:   id,
				Inner: i.Label,
			},
			&h.Div{
				Class: "controls",
				Inner: &h.Label{
					Class: "checkbox",
					Inner: &h.Frag{
						&h.Input{
							Type:    "checkbox",
							ID:      id,
							Name:    i.Name,
							Checked: i.Checked,
							Value:   "1",
						},
						i.Description,
					},
				},
			},
		},
	}, nil
}

type contextEditor struct {
	Context *context.Context
	Example *examples.Example
}

func (e *contextEditor) HTML() (h.HTML, error) {
	if !e.Context.IsEmployee {
		return nil, nil
	}
	return &h.Div{
		Class: "well form-horizontal",
		Inner: &h.Frag{
			&textInput{
				Label: h.String("Application ID"),
				Name:  "appid",
				Value: e.Context.AppID,
			},
			&checkboxInput{
				Label:       h.String("Init"),
				Name:        "init",
				Checked:     e.Context.Init,
				Description: h.String("Automatically initialize SDK."),
			},
			&checkboxInput{
				Label:       h.String("Status"),
				Name:        "status",
				Checked:     e.Context.Status,
				Description: h.String("Automatically trigger status ping."),
			},
			&checkboxInput{
				Label:       h.String("Channel"),
				Name:        "channel",
				Checked:     e.Context.UseChannel,
				Description: h.String("Specify explicit XD channel."),
			},
			&checkboxInput{
				Label:       h.String("Frictionless Requests"),
				Name:        "frictionlessRequests",
				Checked:     e.Context.FrictionlessRequests,
				Description: h.String("Enable frictionless requests."),
			},
			&h.Div{
				Class: "form-actions",
				Inner: &h.Frag{
					&h.Button{
						Type:  "submit",
						Class: "btn btn-primary",
						Inner: h.String("Update"),
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
	frag := &h.Frag{}
	foundSelected := false
	selected := false
	for title, value := range envOptions {
		selected = e.Context.Env == value
		if selected {
			foundSelected = true
		}
		ctxCopy := e.Context.Copy()
		ctxCopy.Env = value
		frag.Append(&h.Option{
			Inner:    h.String(title),
			Selected: selected,
			Value:    value,
			Data: map[string]interface{}{
				"url": ctxCopy.ViewURL(e.Example.URL),
			},
		})
	}
	if !foundSelected {
		frag.Append(&h.Option{
			Inner:    h.String(e.Context.Env),
			Selected: true,
			Value:    e.Context.Env,
			Data: map[string]interface{}{
				"url": e.Context.ViewURL(e.Example.URL),
			},
		})
	}
	return &h.Select{
		ID:    "rell-env",
		Name:  "env",
		Inner: frag,
	}, nil
}
