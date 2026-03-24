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
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"

	"github.com/daaku/ctxerr"
	"github.com/daaku/go.fburl"
	"github.com/daaku/go.h"
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
		DB:      a.ExampleStore.DB,
	})
	return err
}

// sortedCategories returns the DB categories sorted by name, excluding hidden ones.
func sortedCategories(db *examples.DB) []*examples.Category {
	var cats []*examples.Category
	for _, cat := range db.Category {
		if !cat.Hidden {
			cats = append(cats, cat)
		}
	}
	sort.Slice(cats, func(i, j int) bool {
		return cats[i].Name < cats[j].Name
	})
	return cats
}

// headerBar renders the top navigation bar.
type headerBar struct {
	Context context.Context
	Env     *rellenv.Env
	Example *examples.Example
}

func (hb *headerBar) HTML(ctx context.Context) (h.HTML, error) {
	authControls := h.Frag{
		&h.Div{
			Class: "fb-login-button",
			ID:    "fb-login-plugin",
			Data: map[string]interface{}{
				"size":              "medium",
				"button-type":       "continue_with",
				"use-continue-as":   "true",
				"auto-logout-link":  "true",
			},
		},
		&h.Button{
			ID:    "fb-login-custom",
			Class: "btn btn-login",
			Inner: h.String("Log in with Facebook"),
		},
		&h.Span{Class: "auth-label", Inner: h.String("Status:")},
		&h.Span{ID: "auth-status", Class: "auth-badge", Inner: h.String("waiting")},
		&h.A{ID: "rell-disconnect", Class: "btn-link", Inner: h.String("Disconnect")},
	}

	headerRight := h.Frag{
		&h.Button{ID: "settings-toggle", Class: "btn-icon", Inner: h.Unsafe("&#9881;")},
		&h.Div{Class: "auth-controls", Inner: authControls},
	}

	examplesURL := "/"
	if hb.Env != nil {
		examplesURL = hb.Env.URL("/examples/").String()
	}

	headerLeft := h.Frag{
		&h.A{
			Class: "logo",
			HREF:  "/",
			Inner: h.Frag{
				h.String("fbrell"),
				&h.Span{Class: "logo-accent", Inner: h.String(">_")},
			},
		},
		&h.Button{
			ID:    "sidebar-toggle",
			Class: "btn-icon sidebar-toggle",
			Inner: h.Unsafe("&#9776;"),
		},
		&h.Div{
			Class: "header-nav",
			Inner: h.Frag{
				&h.A{
					Class: "nav-link",
					HREF:  examplesURL,
					Inner: h.String("Examples"),
				},
				&h.A{
					Class:  "nav-link",
					HREF:   "https://developers.facebook.com/docs/javascript/",
					Target: "_blank",
					Inner:  h.String("Docs"),
				},
			},
		},
	}

	if rellenv.IsEmployee(hb.Context) && hb.Example != nil {
		fbEnv := rellenv.FbEnv(hb.Context)
		title := envOptions[fbEnv]
		if title == "" {
			title = fbEnv
		}

		var envItems h.Frag
		for _, pair := range sortutil.StringMapByValue(envOptions) {
			if fbEnv == pair.Key {
				continue
			}
			ctxCopy := hb.Env.Copy()
			ctxCopy.Env = pair.Key
			envItems = append(envItems, &h.A{
				Class:  "env-option",
				Target: "_top",
				HREF:   ctxCopy.ViewURL(hb.Example.URL),
				Inner:  h.String(pair.Value),
			})
		}

		envSelector := &h.Div{
			Class: "env-selector",
			Inner: h.Frag{
				&h.Button{
					Class: "btn btn-env",
					Inner: h.String(title),
				},
				&h.Div{
					Class: "env-dropdown",
					Inner: envItems,
				},
				h.HiddenInputs(url.Values{
					"server": []string{fbEnv},
				}),
			},
		}

		headerRight = append(h.Frag{envSelector}, headerRight...)
	}

	return &h.Header{
		Class: "header",
		Inner: h.Frag{
			&h.Div{
				Class: "header-left",
				Inner: headerLeft,
			},
			&h.Div{
				Class: "header-right",
				Inner: headerRight,
			},
		},
	}, nil
}

// sidebarNav renders the sidebar with example tree navigation.
type sidebarNav struct {
	Context context.Context
	Env     *rellenv.Env
	DB      *examples.DB
}

func (s *sidebarNav) HTML(ctx context.Context) (h.HTML, error) {
	cats := sortedCategories(s.DB)
	var categories h.Frag
	for _, category := range cats {
		var items h.Frag
		for _, example := range category.Example {
			items = append(items, &h.A{
				Class: "sidebar-item",
				HREF:  s.Env.URL(example.URL).String(),
				Inner: h.String(example.Name),
			})
		}
		categories = append(categories, &h.Div{
			Class: "sidebar-category",
			Inner: h.Frag{
				&h.Div{
					Class: "sidebar-category-header",
					Inner: h.Frag{
						&h.Span{Class: "sidebar-toggle", Inner: h.Unsafe("&#9660;")},
						&h.Span{Class: "sidebar-category-name", Inner: h.String(category.Name)},
					},
				},
				&h.Div{
					Class: "sidebar-category-items",
					Inner: items,
				},
			},
		})
	}
	return &h.Div{
		ID:    "sidebar",
		Class: "sidebar",
		Inner: h.Frag{
			&h.Div{
				Class: "sidebar-header",
				Inner: &h.Input{
					ID:          "sidebar-search",
					Class:       "sidebar-search",
					Type:        "text",
					Placeholder: "Search examples...",
				},
			},
			&h.Div{
				ID:    "sidebar-tree",
				Class: "sidebar-tree",
				Inner: categories,
			},
		},
	}, nil
}

// logPanel renders the log column with filter and action buttons.
type logPanel struct{}


func (l *logPanel) HTML(ctx context.Context) (h.HTML, error) {
	return &h.Div{
		Class: "log-column",
		Inner: h.Frag{
			&h.Div{
				Class: "log-header",
				Inner: h.Frag{
					&h.Span{Class: "log-title", Inner: h.String("Log")},
					&h.Span{ID: "log-error-count", Class: "log-count", Style: "display:none"},
					&h.Div{
						Class: "log-actions",
						Inner: h.Frag{
							&h.Button{ID: "rell-log-clear", Class: "btn-icon", Inner: h.Unsafe("&#128465;")},
							&h.Button{ID: "rell-log-copy", Class: "btn-icon", Inner: h.Unsafe("&#128203;")},
						},
					},
				},
			},
			&h.Div{ID: "log", Class: "log-entries"},
		},
	}, nil
}

// statusBar renders the bottom status bar.
type statusBar struct{}

func (s *statusBar) HTML(ctx context.Context) (h.HTML, error) {
	return &h.Div{
		Class: "status-bar",
		Inner: h.Frag{
			&h.Span{ID: "sdk-status", Class: "status-item", Inner: h.Frag{
				&h.Span{Class: "status-dot status-dot-warning"},
				h.String(" Loading SDK..."),
			}},
			&h.Span{ID: "sdk-version", Class: "status-item"},
			&h.Span{ID: "app-id-display", Class: "status-item"},
			&h.Button{ID: "theme-toggle", Class: "btn-icon status-item", Inner: h.Unsafe("&#9788;")},
			&h.A{
				Class:  "status-item status-link",
				HREF:   "https://github.com/fbsamples/fbrell",
				Target: "_blank",
				Inner:  h.String("GitHub"),
			},
		},
	}, nil
}

// settingsDrawer renders the settings panel as a slide-out drawer.
type settingsDrawer struct {
	Context context.Context
	Env     *rellenv.Env
	Example *examples.Example
}

func (s *settingsDrawer) HTML(ctx context.Context) (h.HTML, error) {
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

	return &h.Div{
		ID:    "settings-drawer",
		Class: "settings-drawer",
		Inner: h.Frag{
			&h.Div{
				Class: "drawer-header",
				Inner: h.Frag{
					&h.Span{Inner: h.String("Settings")},
					&h.Button{
						ID:    "settings-close",
						Class: "btn-icon",
						Inner: h.Unsafe("&#10005;"),
					},
				},
			},
			&h.Div{
				Class: "drawer-body",
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
					&settingsCheckbox{
						Label:   "Custom Login Button",
						Name:    "customLogin",
						Checked: s.Env.CustomLogin,
					},
				},
			},
			&h.Div{
				Class: "drawer-footer",
				Inner: h.Frag{
					&h.Div{
						ID:    "rell-url-preview",
						Class: "rell-url-preview",
						Inner: &h.Small{
							Inner: h.String(s.Env.URL(s.Example.URL).String()),
						},
					},
					&h.Button{
						ID:    "rell-settings-update",
						Class: "btn btn-primary",
						Inner: h.String("Update"),
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
		Class: "setting-group",
		Inner: h.Frag{
			&h.Label{
				Class: "setting-label",
				Inner: h.String(f.Label),
			},
			&h.Input{
				Class:       "rell-setting",
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
		Class: "setting-group",
		Inner: h.Frag{
			&h.Label{
				Class: "setting-label",
				Inner: h.String(s.Label),
			},
			&h.Select{
				Class: "rell-setting",
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
		Class: "setting-group setting-group-checkbox",
		Inner: &h.Label{
			Class: "setting-label",
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

type page struct {
	Writer  http.ResponseWriter
	Request *http.Request
	Context context.Context
	Env     *rellenv.Env
	Static  *static.Handler
	Example *examples.Example
	DB      *examples.DB
}

func (p *page) HTML(ctx context.Context) (h.HTML, error) {
	return &view.Page{
		Title: p.Example.Title,
		Class: "main",
		Body: &h.Div{
			Class: "app",
			Inner: h.Frag{
				&headerBar{
					Context: p.Context,
					Env:     p.Env,
					Example: p.Example,
				},
				&h.Div{
					Class: "main-layout",
					Inner: h.Frag{
						&sidebarNav{
							Context: p.Context,
							Env:     p.Env,
							DB:      p.DB,
						},
						&h.Div{
							Class: "mobile-tab-bar",
							Inner: h.Frag{
								&h.Button{Class: "mobile-tab active", Data: map[string]interface{}{"tab": "editor"}, Inner: h.String("Editor")},
								&h.Button{Class: "mobile-tab", Data: map[string]interface{}{"tab": "output"}, Inner: h.String("Output")},
								&h.Button{Class: "mobile-tab", Data: map[string]interface{}{"tab": "log"}, Inner: h.String("Log")},
							},
						},
						&h.Div{
							Class: "editor-column",
							Inner: h.Frag{
								&h.Div{
									Class: "editor-toolbar",
									Inner: h.Frag{
										&h.Div{
											Class: "toolbar-left",
											Inner: &h.Span{
												Class: "example-title",
												Inner: h.String(p.Example.Title),
											},
										},
										&h.Div{
											Class: "toolbar-right",
											Inner: &h.Button{
												ID:    "rell-run-code",
												Class: "btn btn-primary",
												Inner: h.Unsafe("&#9654; Run"),
											},
										},
									},
								},
								&h.Div{
									ID:    "editor-pane",
									Class: "editor-pane",
									Inner: &h.Textarea{
										ID:   "jscode",
										Name: "code",
										Inner: &exampleContent{
											Context: p.Context,
											Env:     p.Env,
											Example: p.Example,
										},
									},
								},
								&h.Div{ID: "resize-v", Class: "resize-handle resize-handle-v"},
								&h.Div{
									ID:    "jsroot",
									Class: "output-pane",
									Inner: &h.Div{
										Class: "empty-state",
										Inner: h.String("// Output will appear here. Press Cmd+Enter to run."),
									},
								},
							},
						},
						&h.Div{ID: "resize-h", Class: "resize-handle resize-handle-h"},
						&logPanel{},
					},
				},
				&statusBar{},
				&h.Div{ID: "sidebar-overlay", Class: "sidebar-overlay"},
				&h.Div{ID: "settings-overlay", Class: "settings-overlay"},
				&settingsDrawer{
					Context: p.Context,
					Env:     p.Env,
					Example: p.Example,
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

type examplesList struct {
	Context context.Context
	Env     *rellenv.Env
	DB      *examples.DB
	Static  *static.Handler
}

func (l *examplesList) HTML(ctx context.Context) (h.HTML, error) {
	cats := sortedCategories(l.DB)
	var cards h.Frag
	for _, category := range cats {
		var links h.Frag
		for _, example := range category.Example {
			links = append(links, &h.A{
				Class: "example-link",
				HREF:  l.Env.URL(example.URL).String(),
				Inner: &h.Span{
					Class: "example-name",
					Inner: h.String(example.Name),
				},
			})
		}
		cards = append(cards, &h.Div{
			Class: "example-category-card",
			Inner: h.Frag{
				&h.Div{
					Class: "category-header",
					Inner: h.Frag{
						&h.Span{Class: "category-icon", Inner: h.String(categoryIcon(category.Name))},
						&h.H2{Inner: h.String(category.Name)},
					},
				},
				&h.Div{
					Class: "category-examples",
					Inner: links,
				},
			},
		})
	}

	return &view.Page{
		Title: "Examples",
		Class: "examples",
		Body: &h.Div{
			Class: "app",
			Inner: h.Frag{
				&headerBar{
					Context: l.Context,
					Env:     l.Env,
				},
				&h.Div{
					Class: "examples-page",
					Inner: h.Frag{
						&h.Div{
							Class: "examples-header",
							Inner: h.Frag{
								&h.H1{Inner: h.String("Examples")},
								&h.Input{
									ID:          "examples-search",
									Class:       "examples-search",
									Type:        "text",
									Placeholder: "Search examples...",
								},
							},
						},
						&h.Div{
							Class: "examples-grid",
							Inner: cards,
						},
					},
				},
				&statusBar{},
			},
		},
	}, nil
}

func categoryIcon(name string) string {
	switch name {
	case "Facebook Login":
		return "\xf0\x9f\x94\x91" // key
	case "Graph API":
		return "\xf0\x9f\x93\x8a" // chart
	case "Sharing":
		return "\xf0\x9f\x94\x97" // link
	case "auth":
		return "\xf0\x9f\x94\x90" // lock
	default:
		return "\xf0\x9f\x93\x9d" // memo
	}
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
