// Package viewexamples provides the various views on Rell examples.
package viewexamples

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/ctxerr"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.errcode"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.fburl"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.h"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.h.ui"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.htmlwriter"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.static"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.xsrf"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/sortutil"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/facebookgo/counting"
	"github.com/daaku/rell/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/daaku/rell/examples"
	"github.com/daaku/rell/rellenv"
	"github.com/daaku/rell/view"
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
		"sb":             "Sandbox",
	}
	viewModeOptions = map[string]string{
		rellenv.Website: "Website",
		rellenv.PageTab: "Page Tab",
		rellenv.Canvas:  "Canvas",
	}
	errTokenMismatch = errcode.New(http.StatusForbidden, "Token mismatch.")
	errSaveDisabled  = errcode.New(http.StatusForbidden, "Save disallowed.")
)

type Handler struct {
	ExampleStore *examples.Store
	Static       *static.Handler
	Xsrf         *xsrf.Provider
}

// Parse the Context and an Example.
func (h *Handler) parse(ctx context.Context, r *http.Request) (*rellenv.Env, *examples.Example, error) {
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

func (a *Handler) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	context, err := rellenv.FromContext(ctx)
	if err != nil {
		return err
	}
	h.WriteResponse(w, r, &examplesList{
		Context: context,
		Static:  a.Static,
		DB:      a.ExampleStore.DB,
	})
	return nil
}

func (a *Handler) PostSaved(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	c, err := rellenv.FromContext(ctx)
	if err != nil {
		return err
	}
	if !c.IsEmployee {
		return ctxerr.Wrap(ctx, errSaveDisabled)
	}
	if !a.Xsrf.Validate(r.FormValue(paramName), w, r, savedPath) {
		return ctxerr.Wrap(ctx, errTokenMismatch)
	}
	content := strings.TrimSpace(r.FormValue("code"))
	content = strings.Replace(content, "\x13", "", -1) // remove CR
	id := examples.ContentID(content)
	db := a.ExampleStore.DB
	example, ok := db.Reverse[id]
	if ok {
		http.Redirect(w, r, c.ViewURL(example.URL), 302)
		return nil
	}
	err = a.ExampleStore.Save(id, content)
	if err != nil {
		return err
	}
	http.Redirect(w, r, c.ViewURL(savedPath+id), 302)
	return nil
}

func (a *Handler) GetSaved(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	context, example, err := a.parse(ctx, r)
	if err != nil {
		return err
	}
	h.WriteResponse(w, r, &page{
		Writer:  w,
		Request: r,
		Context: context,
		Static:  a.Static,
		Example: example,
		Xsrf:    a.Xsrf,
	})
	return nil
}

func (a *Handler) Example(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	context, example, err := a.parse(ctx, r)
	if err != nil {
		return err
	}
	h.WriteResponse(w, r, &page{
		Writer:  w,
		Request: r,
		Context: context,
		Static:  a.Static,
		Example: example,
		Xsrf:    a.Xsrf,
	})
	return nil
}

type page struct {
	Writer  http.ResponseWriter
	Request *http.Request
	Context *rellenv.Env
	Static  *static.Handler
	Example *examples.Example
	Xsrf    *xsrf.Provider
}

func (p *page) HTML() (h.HTML, error) {
	return &view.Page{
		Static: p.Static,
		Title:  p.Example.Title,
		Class:  "main",
		Body: &h.Div{
			Class: "container-fluid",
			Inner: &h.Frag{
				&h.Form{
					Action: savedPath,
					Method: h.Post,
					Target: "_top",
					Inner: &h.Frag{
						h.HiddenInputs(url.Values{
							paramName: []string{p.Xsrf.Token(p.Writer, p.Request, savedPath)},
						}),
						&h.Div{
							Class: "row-fluid",
							Inner: &h.Frag{
								&h.Div{
									Class: "span8",
									Inner: &h.Frag{
										&editorTop{Context: p.Context, Example: p.Example},
										&editorArea{
											Context: p.Context,
											Example: p.Example,
										},
										&editorBottom{Context: p.Context, Example: p.Example},
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
				&h.Div{
					Class: "row-fluid",
					Inner: &h.Div{
						Class: "span12",
						Inner: &editorOutput{},
					},
				},
				&JsInit{
					Context: p.Context,
					Example: p.Example,
				},
			},
		},
	}, nil
}

type editorTop struct {
	Context *rellenv.Env
	Example *examples.Example
}

func (e *editorTop) HTML() (h.HTML, error) {
	left := &h.Frag{
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

	if e.Context.IsEmployee {
		return &h.Div{
			Class: "row-fluid form-inline",
			Inner: &h.Frag{
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
							Example: e.Example,
						},
					},
				},
			},
		}, nil
	}
	return &h.Div{
		Class: "row-fluid form-inline",
		Inner: &h.Frag{
			&h.Div{
				Class: "span12",
				Inner: left,
			},
		},
	}, nil
}

type editorArea struct {
	Context *rellenv.Env
	Example *examples.Example
}

func (e *editorArea) HTML() (h.HTML, error) {
	return &h.Div{
		Class: "row-fluid",
		Inner: &h.Textarea{
			ID:   "jscode",
			Name: "code",
			Inner: &exampleContent{
				Context: e.Context,
				Example: e.Example,
			},
		},
	}, nil
}

type viewModeDropdown struct {
	Context *rellenv.Env
	Example *examples.Example
}

func (d *viewModeDropdown) HTML() (h.HTML, error) {
	return &h.Div{
		Class: "btn-group",
		Inner: &h.Frag{
			&h.Button{
				Class: "btn",
				Inner: &h.Frag{
					&h.I{Class: "icon-eye-open"},
					h.String(" "),
					h.String(viewModeOptions[d.Context.ViewMode]),
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
				Inner: &h.Frag{
					&h.Li{
						Inner: &h.A{
							Inner:  h.String(viewModeOptions[rellenv.Website]),
							Target: "_top",
							HREF:   d.Context.URL(d.Example.URL).String(),
						},
					},
					&h.Li{
						Inner: &h.A{
							Inner:  h.String(viewModeOptions[rellenv.Canvas]),
							Target: "_top",
							HREF:   d.Context.CanvasURL(d.Example.URL),
						},
					},
					&h.Li{
						Inner: &h.A{
							Inner:  h.String(viewModeOptions[rellenv.PageTab]),
							Target: "_top",
							HREF:   d.Context.PageTabURL(d.Example.URL),
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
					Value: d.Context.ViewMode,
				},
			},
		},
	}, nil
}

type editorBottom struct {
	Context *rellenv.Env
	Example *examples.Example
}

func (e *editorBottom) HTML() (h.HTML, error) {
	runButton := &h.A{
		ID:    "rell-run-code",
		Class: "btn btn-primary",
		Inner: &h.Frag{
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
	var saveButton h.HTML
	if e.Context.IsEmployee {
		saveButton = &h.Frag{
			h.String(" "),
			&h.Div{
				Class: "btn-group",
				Inner: &h.Button{
					Class: "btn",
					Type:  "submit",
					Inner: &h.Frag{
						&h.I{Class: "icon-file"},
						h.String(" Save Code"),
					},
				},
			},
		}
	}
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
						saveButton,
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
	Context *rellenv.Env
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
				Tooltip:    "Make sure the base domain in the application settings for the specified ID allows fbrell.com.",
			},
			&ui.ToggleGroup{
				Inner: &h.Frag{
					&ui.ToggleItem{
						Name:        "init",
						Checked:     e.Context.Init,
						Description: h.String("Automatically initialize SDK."),
						Tooltip:     "This controls if FB.init() is automatically called. If off, you'll need to call it in your code.",
					},
					&ui.ToggleItem{
						Name:        "status",
						Checked:     e.Context.Status,
						Description: h.String("Automatically trigger status ping."),
						Tooltip:     "This controls the \"status\" parameter to FB.init.",
					},
					&ui.ToggleItem{
						Name:        "frictionlessRequests",
						Checked:     e.Context.FrictionlessRequests,
						Description: h.String("Enable frictionless requests."),
						Tooltip:     "This controls the \"frictionlessRequests\" parameter to FB.init.",
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
	Context *rellenv.Env
	DB      *examples.DB
	Static  *static.Handler
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
		Static: l.Static,
		Title:  "Examples",
		Class:  "examples",
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
	Context  *rellenv.Env
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
	Context *rellenv.Env
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
	for _, pair := range sortutil.StringMapByValue(envOptions) {
		if e.Context.Env == pair.Key {
			continue
		}
		ctxCopy := e.Context.Copy()
		ctxCopy.Env = pair.Key
		frag.Append(&h.Li{
			Inner: &h.A{
				Inner:  h.String(pair.Value),
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
				Inner: &h.Frag{
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
	Context *rellenv.Env
	Example *examples.Example
}

func (c *exampleContent) HTML() (h.HTML, error) {
	return c, fmt.Errorf("exampleContent.HTML is a dangerous primitive")
}

// Renders the example content including support for context sensitive
// text substitution.
func (c *exampleContent) Write(w io.Writer) (int, error) {
	e := c.Example
	wwwURL := fburl.URL{
		Env: c.Context.Env,
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
			RellFBNS: c.Context.AppNamespace,
			RellURL:  c.Context.AbsoluteURL("/").String(),
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
	Context *rellenv.Env
	Example *examples.Example
}

func (i *JsInit) HTML() (h.HTML, error) {
	encodedContext, err := json.Marshal(i.Context)
	if err != nil {
		return nil, fmt.Errorf("Failed to json.Marshal context: %s", err)
	}
	encodedExample, err := json.Marshal(i.Example)
	if err != nil {
		return nil, fmt.Errorf("Failed to json.Marshal example: %s", err)
	}
	return &h.Frag{
		&h.Script{
			Src:   i.Context.SdkURL(),
			Async: true,
		},
		&h.Script{
			Inner: &h.Frag{
				h.Unsafe("window.rellConfig="),
				h.UnsafeBytes(encodedContext),
				h.Unsafe(";window.rellExample="),
				h.UnsafeBytes(encodedExample),
			},
		},
	}, nil
}
