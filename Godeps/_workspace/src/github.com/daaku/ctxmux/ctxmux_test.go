package ctxmux_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/ctxmux"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/facebookgo/ensure"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/julienschmidt/httprouter"
	"github.com/daaku/rell/Godeps/_workspace/src/golang.org/x/net/context"
)

func TestContextWithNoParams(t *testing.T) {
	var nilParams httprouter.Params
	ensure.DeepEqual(t, ctxmux.ContextParams(context.Background()), nilParams)
}

func TestContextWithFromParams(t *testing.T) {
	p := httprouter.Params{}
	ctx := ctxmux.WithParams(context.Background(), p)
	actual := ctxmux.ContextParams(ctx)
	ensure.DeepEqual(t, actual, p)
}

func TestHTTPHandler(t *testing.T) {
	w := httptest.NewRecorder()
	r := &http.Request{}
	var actualW http.ResponseWriter
	var actualR *http.Request
	h := ctxmux.HTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualW = w
		actualR = r
	}))
	ensure.Nil(t, h(nil, w, r))
	ensure.DeepEqual(t, actualW, w)
	ensure.DeepEqual(t, actualR, r)
}

func TestHTTPHandlerFunc(t *testing.T) {
	w := httptest.NewRecorder()
	r := &http.Request{}
	var actualW http.ResponseWriter
	var actualR *http.Request
	h := ctxmux.HTTPHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualW = w
		actualR = r
	})
	ensure.Nil(t, h(nil, w, r))
	ensure.DeepEqual(t, actualW, w)
	ensure.DeepEqual(t, actualR, r)
}

func TestContextPipeChainSuccess(t *testing.T) {
	const key = int(1)
	const val = int(2)
	p := ctxmux.ContextPipeChain(
		func(ctx context.Context, r *http.Request) (context.Context, error) {
			return context.WithValue(ctx, key, val), nil
		})
	ctx, err := p(context.Background(), nil)
	ensure.Nil(t, err)
	ensure.DeepEqual(t, ctx.Value(key), val)
}

func TestContextPipeChainFailure(t *testing.T) {
	givenErr := errors.New("")
	p := ctxmux.ContextPipeChain(
		func(context.Context, *http.Request) (context.Context, error) {
			return nil, givenErr
		})
	_, err := p(context.Background(), nil)
	ensure.DeepEqual(t, err, givenErr)
}

func TestNewError(t *testing.T) {
	givenErr := errors.New("")
	mux, err := ctxmux.New(
		func(*ctxmux.Mux) error {
			return givenErr
		},
	)
	ensure.True(t, mux == nil)
	ensure.DeepEqual(t, err, givenErr)
}

func TestWrapMethods(t *testing.T) {
	cases := []struct {
		Method   string
		Register func(*ctxmux.Mux, string, ctxmux.Handler)
	}{
		{Method: "HEAD", Register: (*ctxmux.Mux).HEAD},
		{Method: "GET", Register: (*ctxmux.Mux).GET},
		{Method: "POST", Register: (*ctxmux.Mux).POST},
		{Method: "PUT", Register: (*ctxmux.Mux).PUT},
		{Method: "DELETE", Register: (*ctxmux.Mux).DELETE},
		{Method: "PATCH", Register: (*ctxmux.Mux).PATCH},
	}
	const key = int(1)
	const val = int(2)
	body := []byte("body")
	for _, c := range cases {
		mux, err := ctxmux.New(
			ctxmux.MuxContextPipe(ctxmux.ContextPipeChain(
				func(ctx context.Context, r *http.Request) (context.Context, error) {
					return context.WithValue(ctx, key, val), nil
				})),
		)
		ensure.Nil(t, err)
		hw := httptest.NewRecorder()
		hr := &http.Request{
			Method: c.Method,
			URL: &url.URL{
				Path: "/",
			},
		}
		c.Register(mux, hr.URL.Path, func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			ensure.DeepEqual(t, ctx.Value(key), val)
			w.Write(body)
			return nil
		})
		mux.ServeHTTP(hw, hr)
		ensure.DeepEqual(t, hw.Body.Bytes(), body)
	}
}

func TestMuxContextPipeError(t *testing.T) {
	givenErr := errors.New("")
	var actualErr error
	mux, err := ctxmux.New(
		ctxmux.MuxContextPipe(ctxmux.ContextPipeChain(
			func(ctx context.Context, r *http.Request) (context.Context, error) {
				return nil, givenErr
			})),
		ctxmux.MuxErrorHandler(
			func(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
				actualErr = err
			}),
	)
	ensure.Nil(t, err)
	hw := httptest.NewRecorder()
	hr := &http.Request{
		Method: "GET",
		URL: &url.URL{
			Path: "/",
		},
	}
	mux.GET(hr.URL.Path, func(context.Context, http.ResponseWriter, *http.Request) error {
		panic("not reached")
	})
	mux.ServeHTTP(hw, hr)
	ensure.DeepEqual(t, actualErr, givenErr)
}

func TestHandleCustomMethod(t *testing.T) {
	mux, err := ctxmux.New()
	ensure.Nil(t, err)
	const method = "FOO"
	body := []byte("body")
	hw := httptest.NewRecorder()
	hr := &http.Request{
		Method: method,
		URL: &url.URL{
			Path: "/",
		},
	}
	mux.Handler(method, hr.URL.Path, func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		w.Write(body)
		return nil
	})
	mux.ServeHTTP(hw, hr)
	ensure.DeepEqual(t, hw.Body.Bytes(), body)
}

func TestHandlerReturnErr(t *testing.T) {
	givenErr := errors.New("")
	var actualErr error
	mux, err := ctxmux.New(
		ctxmux.MuxErrorHandler(
			func(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
				actualErr = err
			}),
	)
	ensure.Nil(t, err)
	hw := httptest.NewRecorder()
	hr := &http.Request{
		Method: "GET",
		URL: &url.URL{
			Path: "/",
		},
	}
	mux.GET(hr.URL.Path, func(context.Context, http.ResponseWriter, *http.Request) error {
		return givenErr
	})
	mux.ServeHTTP(hw, hr)
	ensure.DeepEqual(t, actualErr, givenErr)
}

func TestHandlerPanic(t *testing.T) {
	var actualPanic interface{}
	mux, err := ctxmux.New(
		ctxmux.MuxPanicHandler(
			func(ctx context.Context, w http.ResponseWriter, r *http.Request, v interface{}) {
				actualPanic = v
			}),
	)
	ensure.Nil(t, err)
	hw := httptest.NewRecorder()
	hr := &http.Request{
		Method: "GET",
		URL: &url.URL{
			Path: "/",
		},
	}
	givenPanic := int(42)
	mux.GET(hr.URL.Path, func(context.Context, http.ResponseWriter, *http.Request) error {
		panic(givenPanic)
	})
	mux.ServeHTTP(hw, hr)
	ensure.DeepEqual(t, actualPanic, givenPanic)
}

func TestHandlerNoPanic(t *testing.T) {
	mux, err := ctxmux.New(
		ctxmux.MuxPanicHandler(
			func(ctx context.Context, w http.ResponseWriter, r *http.Request, v interface{}) {
				panic("not reached")
			}),
	)
	ensure.Nil(t, err)
	hw := httptest.NewRecorder()
	hr := &http.Request{
		Method: "GET",
		URL: &url.URL{
			Path: "/",
		},
	}
	mux.GET(hr.URL.Path, func(context.Context, http.ResponseWriter, *http.Request) error {
		return nil
	})
	mux.ServeHTTP(hw, hr)
}

func TestHandlerNotFound(t *testing.T) {
	var called bool
	mux, err := ctxmux.New(
		ctxmux.MuxNotFoundHandler(
			func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
				called = true
				return nil
			}),
	)
	ensure.Nil(t, err)
	hw := httptest.NewRecorder()
	hr := &http.Request{
		Method: "GET",
		URL: &url.URL{
			Path: "/",
		},
	}
	mux.ServeHTTP(hw, hr)
	ensure.True(t, called)
}
