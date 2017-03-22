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

package view

import (
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/daaku/go.errcode"
	"github.com/daaku/go.h"
)

type ErrorCode interface // HTTP Coded Error.
{
	error
	Code() int
}

// http.Handler for ErrorCode.
type errorCodeHandler struct {
	err ErrorCode
}

// Serve an appropriate response for this error. Currently this means
// HTML or Plain Text.
// TODO(naitik): Extend for JSON.
func (err errorCodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	code := err.err.Code()
	if code == 0 {
		code = http.StatusInternalServerError
	}
	if usePlainText(r) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(code)
		io.Copy(w, strings.NewReader(err.err.Error()))
		w.Write([]byte("\n"))
	} else {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(code)
		page := &Page{
			Body: h.Frag{
				h.String(err.err.Error()),
				&h.Script{Inner: h.Unsafe("window.location.hash = ''")},
			},
		}
		h.Write(context.TODO(), w, page)
	}
}

// Send a error response. If the error also implements http.Handler,
// it will simply be passed control, otherwise the default error
// rendering will be used.
func Error(w http.ResponseWriter, r *http.Request, err error) {
	handler, ok := err.(http.Handler)
	if !ok {
		errCode, ok := err.(ErrorCode)
		if !ok {
			errCode = errcode.Add(500, err)
		}
		handler = errorCodeHandler{
			err: errCode,
		}
	}
	handler.ServeHTTP(w, r)
}

func usePlainText(r *http.Request) bool {
	return strings.Contains(r.UserAgent(), "curl")
}
