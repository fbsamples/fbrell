/**
 * Copyright (c) 2014-present, Facebook, Inc. All rights reserved.
 *
 * You are hereby granted a non-exclusive, worldwide, royalty-free license to use,
 * copy, modify, and distribute this software in source code or binary form for use
 * in connection with the web services and APIs provided by Facebook.
 *
 * As with any software that integrates with the Facebook platform, your use of
 * this software is subject to the Facebook Developer Principles and Policies
 * [http://developers.facebook.com/policy/]. This copyright notice shall be
 * included in all copies or substantial portions of the software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

// Package viewexamples provides the various views on Rell examples.
package viewexamples

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"strconv"
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
	localeOptions = []string{
		"en_US", "af_ZA", "ak_GH", "am_ET", "ar_AR", "as_IN", "ay_BO", "az_AZ",
		"be_BY", "bg_BG", "bm_ML", "bn_IN", "bp_IN", "br_FR", "bs_BA", "bv_DE",
		"ca_ES", "cb_IQ", "ck_US", "co_FR", "cs_CZ", "cx_PH", "cy_GB",
		"da_DK", "de_DE",
		"eh_IN", "el_GR", "em_ZM", "en_GB", "en_IN", "en_OP", "en_PI", "en_UD", "en_XA", "eo_EO",
		"es_CL", "es_CO", "es_ES", "es_LA", "es_MX", "es_VE", "et_EE", "eu_ES",
		"fa_IR", "fb_AA", "fb_AC", "fb_AR", "fb_HA", "fb_HX", "fb_LL", "fb_LS", "fb_LT", "fb_RL", "fb_ZH", "fbt_AC",
		"ff_NG", "fi_FI", "fn_IT", "fo_FO", "fr_CA", "fr_FR", "fv_NG", "fy_NL",
		"ga_IE", "gl_ES", "gn_PY", "gu_IN", "gx_GR",
		"ha_NG", "he_IL", "hi_FB", "hi_IN", "hr_HR", "ht_HT", "hu_HU", "hy_AM",
		"id_ID", "ig_NG", "ik_US", "is_IS", "it_IT", "iu_CA",
		"ja_JP", "ja_KS", "jv_ID",
		"ka_GE", "kk_KZ", "km_KH", "kn_IN", "ko_KR", "ks_IN", "ku_TR", "ky_KG",
		"la_VA", "lg_UG", "li_NL", "ln_CD", "lo_LA", "lr_IT", "lt_LT", "lv_LV",
		"mg_MG", "mi_NZ", "mk_MK", "ml_IN", "mn_MN", "mos_BF", "mr_IN", "ms_MY", "mt_MT", "my_MM",
		"nb_NO", "nd_ZW", "ne_NP", "nh_MX", "nl_BE", "nl_NL", "nn_NO", "nr_ZA", "ns_ZA", "ny_MW",
		"om_ET", "or_IN",
		"pa_IN", "pcm_NG", "pl_PL", "ps_AF", "pt_BR", "pt_PT",
		"qb_DE", "qc_GT", "qe_US", "qk_DZ", "qr_GR", "qs_DE", "qt_US", "qu_PE", "qv_IT", "qz_MM",
		"rm_CH", "rn_BI", "ro_RO", "ru_RU", "rw_RW",
		"sa_IN", "sc_IT", "se_NO", "si_LK", "sk_SK", "sl_SI", "sn_ZW", "so_SO", "sq_AL", "sr_RS", "ss_SZ", "st_ZA", "su_ID", "sv_SE", "sw_KE", "sy_SY", "sz_PL",
		"ta_IN", "te_IN", "tg_TJ", "th_TH", "ti_ET", "tk_TM", "tl_PH", "tl_ST", "tn_BW", "tq_AR", "tpi_PG", "tr_TR", "ts_ZA", "tt_RU", "tz_MA",
		"uk_UA", "ur_PK", "uz_UZ",
		"ve_ZA", "vi_VN",
		"wo_SN",
		"xh_ZA",
		"yi_DE", "yo_NG",
		"zh_CN", "zh_HK", "zh_TW", "zu_ZA", "zz_TR",
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
								&settingsPanel{
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

type settingsPanel struct {
	Context context.Context
	Env     *rellenv.Env
	Example *examples.Example
}

func (s *settingsPanel) HTML(ctx context.Context) (h.HTML, error) {
	appID := strconv.FormatUint(rellenv.FbApp(s.Context).ID(), 10)

	// Build level dropdown options
	levelOptions := h.Frag{}
	for _, lvl := range []string{"debug", "info", "warn", "error"} {
		levelOptions = append(levelOptions, &h.Option{
			Value:    lvl,
			Selected: s.Env.Level() == lvl,
			Inner:    h.String(lvl),
		})
	}

	// Build view-mode dropdown options
	viewModeOpts := h.Frag{}
	for _, mode := range []string{rellenv.Website, rellenv.Canvas, rellenv.PageTab} {
		viewModeOpts = append(viewModeOpts, &h.Option{
			Value:    mode,
			Selected: s.Env.ViewMode == mode,
			Inner:    h.String(viewModeOptions[mode]),
		})
	}

	// Build locale dropdown options
	localeOpts := h.Frag{}
	for _, loc := range localeOptions {
		localeOpts = append(localeOpts, &h.Option{
			Value:    loc,
			Selected: s.Env.Locale() == loc,
			Inner:    h.String(loc),
		})
	}

	return h.Frag{
		&h.Div{
			Class: "row-fluid",
			Style: "margin-bottom: 5px",
			Inner: &h.A{
				ID:    "rell-settings-toggle",
				Class: "btn btn-small",
				Data: map[string]interface{}{
					"toggle": "collapse",
					"target": "#rell-settings",
				},
				Inner: h.Frag{
					&h.I{Class: "icon-cog"},
					h.String(" Settings "),
					&h.Span{Class: "caret"},
				},
			},
		},
		&h.Div{
			ID:    "rell-settings",
			Class: "collapse",
			Inner: &h.Div{
				Class: "well well-small",
				Inner: h.Frag{
					&h.Div{
						Class: "row-fluid",
						Inner: h.Frag{
							&h.Div{
								Class: "span6",
								Inner: h.Frag{
									&settingsField{
										Label:       "App ID",
										Name:        "appid",
										Value:       appID,
										Placeholder: "342526215814610",
									},
									&settingsField{
										Label:       "Server",
										Name:        "server",
										Value:       s.Env.Env,
										Placeholder: "e.g. beta, latest, intern",
									},
									&settingsField{
										Label:       "Version",
										Name:        "version",
										Value:       s.Env.Version,
										Placeholder: "v25.0",
									},
									&settingsSelect{
										Label:   "Locale",
										Name:    "locale",
										Options: localeOpts,
									},
								},
							},
							&h.Div{
								Class: "span6",
								Inner: h.Frag{
									&settingsSelect{
										Label:   "Log Level",
										Name:    "level",
										Options: levelOptions,
									},
									&settingsSelect{
										Label:   "View Mode",
										Name:    "view-mode",
										Options: viewModeOpts,
									},
									&settingsCheckbox{
										Label:   "Auto Init SDK",
										Name:    "init",
										Checked: s.Env.Init,
									},
									&settingsCheckbox{
										Label:   "Auto Status Ping",
										Name:    "status",
										Checked: s.Env.Status,
									},
									&settingsCheckbox{
										Label:   "Frictionless Requests",
										Name:    "frictionlessRequests",
										Checked: s.Env.FrictionlessRequests,
									},
								},
							},
						},
					},
					&h.Div{
						ID:    "rell-url-preview",
						Class: "rell-url-preview",
						Inner: &h.Small{
							Inner: h.String(s.Env.URL(s.Example.URL).String()),
						},
					},
					&h.Div{
						Style: "margin-top: 8px",
						Inner: &h.Button{
							ID:    "rell-settings-update",
							Class: "btn btn-primary btn-small",
							Inner: h.Frag{
								&h.I{Class: "icon-refresh icon-white"},
								h.String(" Update"),
							},
						},
					},
				},
			},
		},
	}, nil
}

type settingsField struct {
	Label       string
	Name        string
	Value       string
	Placeholder string
}

func (f *settingsField) HTML(ctx context.Context) (h.HTML, error) {
	return &h.Div{
		Class: "control-group control-group-sm",
		Inner: h.Frag{
			&h.Label{
				Class: "control-label-sm",
				Inner: h.String(f.Label),
			},
			&h.Input{
				Class:       "rell-setting input-medium",
				Type:        "text",
				Name:        f.Name,
				Value:       f.Value,
				Placeholder: f.Placeholder,
				Data: map[string]interface{}{
					"default": f.Placeholder,
				},
			},
		},
	}, nil
}

type settingsSelect struct {
	Label   string
	Name    string
	Options h.HTML
}

func (s *settingsSelect) HTML(ctx context.Context) (h.HTML, error) {
	return &h.Div{
		Class: "control-group control-group-sm",
		Inner: h.Frag{
			&h.Label{
				Class: "control-label-sm",
				Inner: h.String(s.Label),
			},
			&h.Select{
				Class: "rell-setting input-medium",
				Name:  s.Name,
				Inner: s.Options,
			},
		},
	}, nil
}

type settingsCheckbox struct {
	Label   string
	Name    string
	Checked bool
}

func (c *settingsCheckbox) HTML(ctx context.Context) (h.HTML, error) {
	return &h.Div{
		Class: "control-group control-group-sm",
		Inner: &h.Label{
			Class: "checkbox",
			Inner: h.Frag{
				&h.Input{
					Class:   "rell-setting",
					Type:    "checkbox",
					Name:    c.Name,
					Value:   "true",
					Checked: c.Checked,
					Data: map[string]interface{}{
						"default": "true",
					},
				},
				h.String(" " + c.Label),
			},
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
