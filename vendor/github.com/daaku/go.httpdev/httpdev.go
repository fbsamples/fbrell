// Package httpdev implements some http handlers useful for testing
// and especially while developing http servers.
package httpdev

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/daaku/go.h"
)

var now = time.Now()

// Handler accepts a "duration" query parameter and will sleep and
// respond with a string message.
func Sleep(w http.ResponseWriter, r *http.Request) {
	duration, err := time.ParseDuration(r.FormValue("duration"))
	if err != nil {
		http.Error(w, err.Error(), 400)
	}
	time.Sleep(duration)
	w.Write([]byte(fmt.Sprintf(
		"Started at %s slept for %d nanoseconds with pid %d.\n",
		now,
		duration.Nanoseconds(),
		os.Getpid())))
}

// Prints HTMLized JSON for Browsers & plain text for others.
func HumanJSON(v interface{}, w http.ResponseWriter, r *http.Request) {
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Printf("Error json.MarshalIndent: %s", err)
	}
	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		h.Write(context.Background(), w, &h.Document{
			Inner: h.Frag{
				&h.Head{
					Inner: h.Frag{
						&h.Meta{Charset: "utf-8"},
						&h.Title{h.String("Dump")},
					},
				},
				&h.Body{
					Inner: &h.Pre{
						Inner: h.String(string(out)),
					},
				},
			},
		})
	} else {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write(out)
		w.Write([]byte("\n"))
	}
}

func headerMap(h http.Header) map[string]string {
	r := make(map[string]string)
	for name, value := range h {
		r[name] = value[0]
	}
	return r
}

// Info handler to see a JSON view of some server context.
func Info(context map[string]interface{}, w http.ResponseWriter, r *http.Request) {
	context["request"] = map[string]interface{}{
		"method": r.Method,
		"form":   r.Form,
		"url": map[string]interface{}{
			"path":  r.URL.Path,
			"query": r.URL.RawQuery,
		},
		"headers": headerMap(r.Header),
	}
	HumanJSON(context, w, r)
}
