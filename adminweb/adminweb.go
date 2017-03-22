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

package adminweb

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"path"
	"sync"

	"github.com/daaku/go.httpdev"
	"github.com/daaku/go.trustforward"
	"github.com/daaku/go.viewvar"
)

type Handler struct {
	Forwarded *trustforward.Forwarded
	SkipHTTPS bool
	Path      string

	mux  http.Handler
	once sync.Once
}

var httpsRequired = []byte("https required\n")

// Serve HTTP requests for the admin port.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.SkipHTTPS && h.Forwarded.Scheme(r) != "https" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(httpsRequired))
		return
	}
	h.once.Do(func() {
		root := path.Join("/", h.Path) + "/"
		mux := http.NewServeMux()
		mux.HandleFunc(root+"debug/pprof/", pprof.Index)
		mux.HandleFunc(root+"debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc(root+"debug/pprof/profile", pprof.Profile)
		mux.HandleFunc(root+"debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc(root+"vars/", viewvar.Json)
		mux.HandleFunc(root+"env/", h.envHandler)
		mux.HandleFunc(root+"sleep/", httpdev.Sleep)
		h.mux = mux
	})
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) envHandler(w http.ResponseWriter, r *http.Request) {
	for _, s := range os.Environ() {
		fmt.Fprintln(w, s)
	}
}
