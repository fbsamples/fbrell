package view

import (
	"github.com/nshah/go.errcode"
	"github.com/nshah/go.h"
	"io"
	"log"
	"net/http"
	"strings"
)

// HTTP Coded Error.
type ErrorCode interface {
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
	if code != http.StatusNotFound {
		log.Printf("Error %d: %s %v", code, r.URL, err)
	}
	w.WriteHeader(code)
	if usePlainText(r) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		io.Copy(w, strings.NewReader(err.err.Error()))
		w.Write([]byte("\n"))
	} else {
		page := &Page{
			Body: h.String(err.err.Error()),
		}
		Write(w, r, page)
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
		handler = errorCodeHandler{errCode}
	}
	handler.ServeHTTP(w, r)
}

func usePlainText(r *http.Request) bool {
	return strings.Contains(r.UserAgent(), "curl")
}
