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

// Package viewcontext provides context related HTTP Handlers for Rell.
package viewcontext

import (
	"net/http"
	"runtime"

	"github.com/daaku/go.httpdev"
	"github.com/fbsamples/fbrell/rellenv"
)

var rev string

type Handler struct{}

// Handler for /info/ to see a JSON view of some server context.
func (h *Handler) Info(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	env, err := rellenv.FromContext(ctx)
	if err != nil {
		return err
	}
	info := map[string]interface{}{
		"context":        env,
		"pageTabURL":     env.PageTabURL("/"),
		"canvasURL":      env.CanvasURL("/"),
		"sdkURL":         env.SdkURL(),
		"rev":            rev,
		"runtimeVersion": runtime.Version(),
	}
	httpdev.Info(info, w, r)
	return nil
}
