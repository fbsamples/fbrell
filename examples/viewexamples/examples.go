// Package viewexamples provides the various views on Rell examples.
package viewexamples

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/daaku/ctxerr"
	"github.com/daaku/go.fburl"
	"github.com/daaku/go.h"
	"github.com/daaku/go.h.ui"
	"github.com/daaku/go.htmlwriter"
	"github.com/daaku/go.static"
	"github.com/daaku/sortutil"
	"github.com/facebookgo/counting"
	"github.com/fbsamples/fbrell/examples"
	"github.com/fbsamples/fbrell/rellenv"
	"github.com/fbsamples/fbrell/view"
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
		"sb":             "Sandbox",
	}
	viewModeOptions = map[string]string{
		rellenv.Website: "Website",
		rellenv.PageTab: "Page Tab",
		rellenv.Canvas:  "Canvas",
	}
)

type Handler struct {
	ExampleStore *examples.Store
	Static       *static.Handler
}

// Parse the Env and an Example.
func (h *Handler) parse(r *http.Request) (*rellenv.Env, *examples.Example, error) {
	ctx := r.Context()
	context, err := rellenv.FromContext(ctx)
	if err != nil {
		return nil, nil, err
	}
	example, err := h.ExampleStore.Load(r.URL.Path)
	if err != nil {
		return nil, nil, ctxerr.Wrap(ctx, err)
	}
	return context, example, nil
}

func (a *Handler) List(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	env, err := rellenv.FromContext(ctx)
	if err != nil {
		return err
	}
	_, err = h.Write(ctx, w, &examplesList{
		Context: ctx,
		Env:     env,
		Static:  a.Static,
		DB:      a.ExampleStore.DB,
	})
	return err
}

func (a *Handler) Example(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	env, example, err := a.parse(r)
	if err != nil {
		return err
	}
	_, err = h.Write(ctx, w, &page{
		Writer:  w,
		Request: r,
		Context: ctx,
		Env:     env,
		Static:  a.Static,
		Example: example,
	})
	return err
}

type page struct {
	Writer  http.ResponseWriter
	Request *http.Request
	Context context.Context
	Env     *rellenv.Env
	Static  *static.Handler
	Example *examples.Example
}

func (p *page) HTML(ctx context.Context) (h.HTML, error) {
	return &view.Page{
		Title: p.Example.Title,
		Class: "main",
		Body: &h.Div{
			Class: "container-fluid",
			Inner: h.Frag{
				&h.Div{
					Class: "row-fluid",
					Inner: h.Frag{
						&h.Div{
							Class: "span8",
							Inner: h.Frag{
								&editorTop{
									Context: p.Context,
									Env:     p.Env,
									Example: p.Example,
								},
								&editorArea{
									Context: p.Context,
									Env:     p.Env,
									Example: p.Example,
								},
								&editorBottom{
									Context: p.Context,
									Env:     p.Env,
									Example: p.Example,
								},
							},
						},
						&h.Div{
							Class: "span4",
							Inner: h.Frag{
								&contextEditor{
									Context: p.Context,
									Env:     p.Env,
									Example: p.Example,
								},
								&logContainer{},
							},
						},
					},
				},
				&h.Div{
					Class: "row-fluid",
					Inner: &h.Div{
						Class: "span12",
						Inner: &editorOutput{},
					},
				},
				&JsInit{
					Context: p.Context,
					Env:     p.Env,
					Example: p.Example,
				},
			},
		},
	}, nil
}

type editorTop struct {
	Context context.Context
	Env     *rellenv.Env
	Example *examples.Example
}

func (e *editorTop) HTML(ctx context.Context) (h.HTML, error) {
	left := h.Frag{
		&h.A{
			ID: "rell-login",
			Inner: &h.Span{
				Inner: h.String(" Log In"),
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
	}

	if rellenv.IsEmployee(e.Context) {
		return &h.Div{
			Class: "row-fluid form-inline",
			Inner: h.Frag{
				&h.Div{
					Class: "span8",
					Inner: left,
				},
				&h.Div{
					Class: "span4",
					Inner: &h.Div{
						Class: "pull-right",
						Inner: &envSelector{
							Context: e.Context,
							Env:     e.Env,
							Example: e.Example,
						},
					},
				},
			},
		}, nil
	}
	return &h.Div{
		Class: "row-fluid form-inline",
		Inner: h.Frag{
			&h.Div{
				Class: "span12",
				Inner: left,
			},
		},
	}, nil
}

type editorArea struct {
	Context context.Context
	Env     *rellenv.Env
	Example *examples.Example
}

func (e *editorArea) HTML(ctx context.Context) (h.HTML, error) {
	return &h.Div{
		Class: "row-fluid",
		Inner: &h.Textarea{
			ID:   "jscode",
			Name: "code",
			Inner: &exampleContent{
				Context: e.Context,
				Env:     e.Env,
				Example: e.Example,
			},
		},
	}, nil
}

type viewModeDropdown struct {
	Context context.Context
	Env     *rellenv.Env
	Example *examples.Example
}

func (d *viewModeDropdown) HTML(ctx context.Context) (h.HTML, error) {
	return &h.Div{
		Class: "btn-group",
		Inner: h.Frag{
			&h.Button{
				Class: "btn",
				Inner: h.Frag{
					&h.I{Class: "icon-eye-open"},
					h.String(" "),
					h.String(viewModeOptions[d.Env.ViewMode]),
				},
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
				Inner: h.Frag{
					&h.Li{
						Inner: &h.A{
							Inner:  h.String(viewModeOptions[rellenv.Website]),
							Target: "_top",
							HREF:   d.Env.URL(d.Example.URL).String(),
						},
					},
					&h.Li{
						Inner: &h.A{
							Inner:  h.String(viewModeOptions[rellenv.Canvas]),
							Target: "_top",
							HREF:   d.Env.CanvasURL(d.Example.URL),
						},
					},
					&h.Li{
						Inner: &h.A{
							Inner:  h.String(viewModeOptions[rellenv.PageTab]),
							Target: "_top",
							HREF:   d.Env.PageTabURL(d.Example.URL),
						},
					},
				},
			},
			&h.Div{
				Style: "display:none",
				Inner: &h.Input{
					Type:  "hidden",
					ID:    "rell-view-mode",
					Name:  "view-mode",
					Value: d.Env.ViewMode,
				},
			},
		},
	}, nil
}

type editorBottom struct {
	Context context.Context
	Env     *rellenv.Env
	Example *examples.Example
}

func (e *editorBottom) HTML(ctx context.Context) (h.HTML, error) {
	runButton := &h.A{
		ID:    "rell-run-code",
		Class: "btn btn-primary",
		Inner: h.Frag{
			&h.I{Class: "icon-play icon-white"},
			h.String(" Run Code"),
		},
	}
	if !e.Example.AutoRun {
		runButton.Rel = "popover"
		runButton.Data = map[string]interface{}{
			"title":     "Click to Run",
			"content":   "This example does not run automatically. Click this button to run it.",
			"placement": "top",
			"trigger":   "manual",
		}
	}
	return &h.Div{
		Class: "row-fluid form-inline",
		Inner: h.Frag{
			&h.Strong{
				Class: "span4",
				Inner: &h.A{
					HREF:  e.Env.URL("/examples/").String(),
					Inner: h.String("Examples"),
				},
			},
			&h.Div{
				Class: "span8",
				Inner: &h.Div{
					Class: "btn-toolbar pull-right",
					Inner: h.Frag{
						&viewModeDropdown{
							Context: e.Context,
							Env:     e.Env,
							Example: e.Example,
						},
						h.String(" "),
						&h.Div{
							Class: "btn-group",
							Inner: runButton,
						},
					},
				},
			},
		},
	}, nil
}

type editorOutput struct{}

func (e *editorOutput) HTML(ctx context.Context) (h.HTML, error) {
	return &h.Div{Class: "row-fluid", ID: "jsroot"}, nil
}

type logContainer struct{}

func (e *logContainer) HTML(ctx context.Context) (h.HTML, error) {
	return &h.Div{
		ID: "log-container",
		Inner: h.Frag{
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
	Context context.Context
	Env     *rellenv.Env
	Example *examples.Example
}

func (e *contextEditor) HTML(ctx context.Context) (h.HTML, error) {
	if !rellenv.IsEmployee(e.Context) {
		return h.HiddenInputs(e.Env.Values()), nil
	}
	return &h.Div{
		Class: "well form-horizontal",
		Inner: h.Frag{
			&ui.TextInput{
				Label:      h.String("Application ID"),
				Name:       "appid",
				Value:      rellenv.FbApp(e.Context).ID(),
				InputClass: "input-medium",
				Tooltip:    "Make sure the base domain in the application settings for the specified ID allows fbrell.com.",
			},
			&ui.ToggleGroup{
				Inner: h.Frag{
					&ui.ToggleItem{
						Name:        "init",
						Checked:     e.Env.Init,
						Description: h.String("Automatically initialize SDK."),
						Tooltip:     "This controls if FB.init() is automatically called. If off, you'll need to call it in your code.",
					},
					&ui.ToggleItem{
						Name:        "status",
						Checked:     e.Env.Status,
						Description: h.String("Automatically trigger status ping."),
						Tooltip:     "This controls the \"status\" parameter to FB.init.",
					},
					&ui.ToggleItem{
						Name:        "frictionlessRequests",
						Checked:     e.Env.FrictionlessRequests,
						Description: h.String("Enable frictionless requests."),
						Tooltip:     "This controls the \"frictionlessRequests\" parameter to FB.init.",
					},
				},
			},
			&h.Div{
				Class: "form-actions",
				Inner: h.Frag{
					&h.Button{
						Type:  "submit",
						Class: "btn btn-primary",
						Inner: h.Frag{
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
	Context context.Context
	Env     *rellenv.Env
	DB      *examples.DB
	Static  *static.Handler
}

func (l *examplesList) HTML(ctx context.Context) (h.HTML, error) {
	var categories h.Frag
	for _, category := range l.DB.Category {
		if !category.Hidden {
			categories = append(categories, &exampleCategory{
				Context:  l.Context,
				Env:      l.Env,
				Category: category,
			})
		}
	}
	return &view.Page{
		Title: "Examples",
		Class: "examples",
		Body: &h.Div{
			Class: "container",
			Inner: &h.Div{
				Class: "row",
				Inner: &h.Div{
					Class: "span12",
					Inner: h.Frag{
						&h.H1{Inner: h.String("Examples")},
						categories,
					},
				},
			},
		},
	}, nil
}

type exampleCategory struct {
	Context  context.Context
	Env      *rellenv.Env
	Category *examples.Category
}

func (c *exampleCategory) HTML(ctx context.Context) (h.HTML, error) {
	var li h.Frag
	for _, example := range c.Category.Example {
		li = append(li, &h.Li{
			Inner: &h.A{
				HREF:  c.Env.URL(example.URL).String(),
				Inner: h.String(example.Name),
			},
		})
	}
	return h.Frag{
		&h.H2{Inner: h.String(c.Category.Name)},
		&h.Ul{Inner: li},
	}, nil
}

type envSelector struct {
	Context context.Context
	Env     *rellenv.Env
	Example *examples.Example
}

func (e *envSelector) HTML(ctx context.Context) (h.HTML, error) {
	if !rellenv.IsEmployee(e.Context) {
		return nil, nil
	}
	fbEnv := rellenv.FbEnv(e.Context)
	frag := h.Frag{
		h.HiddenInputs(url.Values{
			"server": []string{fbEnv},
		}),
	}
	for _, pair := range sortutil.StringMapByValue(envOptions) {
		if fbEnv == pair.Key {
			continue
		}
		ctxCopy := e.Env.Copy()
		ctxCopy.Env = pair.Key
		frag = append(frag, &h.Li{
			Inner: &h.A{
				Inner:  h.String(pair.Value),
				Target: "_top",
				HREF:   ctxCopy.ViewURL(e.Example.URL),
			},
		})
	}

	title := envOptions[fbEnv]
	if title == "" {
		title = fbEnv
	}
	return &h.Div{
		Class: "btn-group",
		Inner: h.Frag{
			&h.Button{
				Class: "btn",
				Inner: h.Frag{
					&h.I{Class: "icon-road"},
					h.String(" "),
					h.String(title),
				},
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

type exampleContent struct {
	Context context.Context
	Env     *rellenv.Env
	Example *examples.Example
}

func (c *exampleContent) HTML(ctx context.Context) (h.HTML, error) {
	return c, fmt.Errorf("exampleContent.HTML is a dangerous primitive")
}

// Renders the example content including support for context sensitive
// text substitution.
func (c *exampleContent) Write(ctx context.Context, w io.Writer) (int, error) {
	e := c.Example
	wwwURL := fburl.URL{
		Env: rellenv.FbEnv(c.Context),
	}
	w = htmlwriter.New(w)
	tpl, err := template.New("example-" + e.URL).Parse(string(e.Content))
	if err != nil {
		// if template parsing fails, we ignore it. it's probably malformed html
		return fmt.Fprint(w, e.Content)
	}
	countingW := counting.NewWriter(w)
	err = tpl.Execute(countingW,
		struct {
			Rand     string // a random token
			RellFBNS string // the OG namespace
			RellURL  string // local http://www.fbrell.com/ URL
			WwwURL   string // server specific http://www.facebook.com/ URL
		}{
			Rand:     randString(10),
			RellFBNS: rellenv.FbApp(c.Context).Namespace(),
			RellURL:  c.Env.AbsoluteURL("/").String(),
			WwwURL:   wwwURL.String(),
		})
	if err != nil {
		// if template execution fails, we ignore it. it's probably malformed html
		return fmt.Fprint(w, e.Content)
	}
	return countingW.Count(), err
}

// random string
func randString(length int) string {
	i := make([]byte, length)
	_, err := rand.Read(i)
	if err != nil {
		log.Panicf("failed to generate randString: %s", err)
	}
	return hex.EncodeToString(i)
}

// Represents configuration for initializing the rell module. Sets up a couple
// of globals.
type JsInit struct {
	Context context.Context
	Env     *rellenv.Env
	Example *examples.Example
}

func (i *JsInit) HTML(ctx context.Context) (h.HTML, error) {
	encodedEnv, err := json.Marshal(i.Env)
	if err != nil {
		return nil, fmt.Errorf("Failed to json.Marshal context: %s", err)
	}
	encodedExample, err := json.Marshal(i.Example)
	if err != nil {
		return nil, fmt.Errorf("Failed to json.Marshal example: %s", err)
	}
	return h.Frag{
		&h.Script{
			Src:   i.Env.SdkURL(),
			Async: true,
		},
		&h.Script{
			Inner: h.Frag{
				h.Unsafe("window.rellConfig="),
				h.UnsafeBytes(encodedEnv),
				h.Unsafe(";window.rellExample="),
				h.UnsafeBytes(encodedExample),
			},
		},
	}, nil
}
