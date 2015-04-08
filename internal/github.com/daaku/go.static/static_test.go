package static

import (
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"

	"github.com/daaku/rell/internal/github.com/daaku/go.h"
	"github.com/daaku/rell/internal/github.com/facebookgo/ensure"
)

func ensureDisableCaching(t testing.TB, h http.Header) {
	ensure.DeepEqual(t, h.Get("Cache-Control"), "no-cache")
	ensure.DeepEqual(t, h.Get("Pragma"), "no-cache")
}

func TestDisableCaching(t *testing.T) {
	w := httptest.NewRecorder()
	disableCaching(w)
	ensureDisableCaching(t, w.Header())
}

func TestNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	notFound(w)
	ensureDisableCaching(t, w.Header())
	ensure.DeepEqual(t, w.Code, http.StatusNotFound)
}

func TestBadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	badRequest(w)
	ensureDisableCaching(t, w.Header())
	ensure.DeepEqual(t, w.Code, http.StatusBadRequest)
}

func TestErrInvalidURL(t *testing.T) {
	ensure.DeepEqual(t, errInvalidURL("foo").Error(), `static: invalid URL "foo"`)
}

func TestEncodeSingle(t *testing.T) {
	files := []*file{{Name: "foo", Hash: "bar"}}
	value, err := encode(files)
	ensure.Nil(t, err)
	ensure.DeepEqual(t, value, "W1siZm9vIiwiYmFyIl1d")
}

func TestEncodeMultiple(t *testing.T) {
	files := []*file{
		{Name: "foo1", Hash: "bar1"},
		{Name: "foo2", Hash: "bar2"},
	}
	value, err := encode(files)
	ensure.Nil(t, err)
	ensure.DeepEqual(t, value, "W1siZm9vMSIsImJhcjEiXSxbImZvbzIiLCJiYXIyIl1d")
}

func TestDecodeSingle(t *testing.T) {
	value, err := decode("W1siZm9vIiwiYmFyIl1d")
	ensure.Nil(t, err)
	ensure.DeepEqual(t, value, []*file{{Name: "foo", Hash: "bar"}})
}

func TestDecodeMultiple(t *testing.T) {
	value, err := decode("W1siZm9vMSIsImJhcjEiXSxbImZvbzIiLCJiYXIyIl1d")
	ensure.Nil(t, err)
	ensure.DeepEqual(t, value, []*file{
		{Name: "foo1", Hash: "bar1"},
		{Name: "foo2", Hash: "bar2"},
	})
}

func TestDecodeInvalidBase64(t *testing.T) {
	value, err := decode("#")
	ensure.True(t, value == nil)
	ensure.Err(t, err, regexp.MustCompile(`static: invalid URL "#"`))
}

func TestDecodeInvalidJSON(t *testing.T) {
	encoded := base64.URLEncoding.EncodeToString([]byte("x"))
	value, err := decode(encoded)
	ensure.True(t, value == nil)
	ensure.Err(t, err, regexp.MustCompile(`static: invalid URL "eA=="`))
}

func TestDecodeNotTwoParts(t *testing.T) {
	encoded := base64.URLEncoding.EncodeToString([]byte("[[]]"))
	value, err := decode(encoded)
	ensure.True(t, value == nil)
	ensure.Err(t, err, regexp.MustCompile(`static: invalid URL "W1tdXQ=="`))
}

type funcBox func(name string) ([]byte, error)

func (f funcBox) Bytes(name string) ([]byte, error) {
	return f(name)
}

func TestLoadFromBox(t *testing.T) {
	const magic = "foo"
	h := Handler{
		Box: funcBox(func(name string) ([]byte, error) {
			ensure.DeepEqual(t, name, magic)
			return []byte(magic), nil
		}),
	}
	f, err := h.load(magic)
	ensure.Nil(t, err)
	ensure.DeepEqual(t, f, &file{
		Name:    magic,
		Content: []byte(magic),
		Hash:    "acbd18db",
	})
}

func TestLoadFromCache(t *testing.T) {
	const magic = "foo"
	h := Handler{
		Box: funcBox(func(name string) ([]byte, error) {
			ensure.DeepEqual(t, name, magic)
			return []byte(magic), nil
		}),
	}
	f, err := h.load(magic)
	ensure.Nil(t, err)
	ensure.DeepEqual(t, f, &file{
		Name:    magic,
		Content: []byte(magic),
		Hash:    "acbd18db",
	})
	h.Box = nil
	f2, err := h.load(magic)
	ensure.Nil(t, err)
	ensure.True(t, f == f2)
}

func TestLoadFromBoxError(t *testing.T) {
	const msg = "foo"
	h := Handler{
		Box: funcBox(func(name string) ([]byte, error) {
			return nil, errors.New(msg)
		}),
	}
	f, err := h.load("baz")
	ensure.True(t, f == nil)
	ensure.Err(t, err, regexp.MustCompile(msg))
}

func TestCombinedURLNoNames(t *testing.T) {
	var h Handler
	v, err := h.URL()
	ensure.DeepEqual(t, v, "")
	ensure.Err(t, err, regexp.MustCompile("zero names given"))
}

func TestCombinedURLBoxError(t *testing.T) {
	const msg = "foo"
	h := Handler{
		Box: funcBox(func(name string) ([]byte, error) {
			return nil, errors.New(msg)
		}),
	}
	v, err := h.URL("baz")
	ensure.DeepEqual(t, v, "")
	ensure.Err(t, err, regexp.MustCompile(msg))
}

func TestCombinedURLMultiple(t *testing.T) {
	contents := [][]byte{
		[]byte("foo"),
		[]byte("bar"),
	}
	var count int
	h := Handler{
		Box: funcBox(func(name string) ([]byte, error) {
			defer func() { count++ }()
			return contents[count], nil
		}),
	}
	v, err := h.URL("n1", "n2")
	ensure.Nil(t, err)
	ensure.DeepEqual(t, v, "W1sibjEiLCJhY2JkMThkYiJdLFsibjIiLCIzN2I1MWQxOSJdXQ==")
}

func TestCombinedURLExt(t *testing.T) {
	contents := [][]byte{
		[]byte("foo"),
		[]byte("bar"),
	}
	var count int
	h := Handler{
		Box: funcBox(func(name string) ([]byte, error) {
			defer func() { count++ }()
			return contents[count], nil
		}),
	}
	v, err := h.URL("n1.js", "n2")
	ensure.Nil(t, err)
	ensure.DeepEqual(t, v, "W1sibjEuanMiLCJhY2JkMThkYiJdLFsibjIiLCIzN2I1MWQxOSJdXQ==.js")
}

func TestServeCombinedURLWithExt(t *testing.T) {
	contents := [][]byte{
		[]byte("foo"),
		[]byte("bar"),
	}
	var count int
	h := Handler{
		Path: "/",
		Box: funcBox(func(name string) ([]byte, error) {
			defer func() { count++ }()
			return contents[count], nil
		}),
	}
	v, err := h.URL("n1.js", "n2")
	ensure.Nil(t, err)
	ensure.DeepEqual(t, v, "/W1sibjEuanMiLCJhY2JkMThkYiJdLFsibjIiLCIzN2I1MWQxOSJdXQ==.js")
	w := httptest.NewRecorder()
	r := &http.Request{
		URL: &url.URL{
			Path: v,
		},
	}
	h.ServeHTTP(w, r)
	ensure.DeepEqual(t, w.Code, http.StatusOK)
	ensure.DeepEqual(t, w.Body.String(), "foobar")
	ensure.DeepEqual(t, w.Header(), http.Header{
		"Content-Length": []string{"6"},
		"Cache-Control":  []string{"public, max-age=315360000"},
		"Content-Type":   []string{"application/javascript"},
	})
}

func TestServeIncorrectPath(t *testing.T) {
	h := Handler{Path: "/foo"}
	w := httptest.NewRecorder()
	r := &http.Request{
		URL: &url.URL{
			Path: "/bar",
		},
	}
	h.ServeHTTP(w, r)
	ensure.DeepEqual(t, w.Code, http.StatusNotFound)
	ensureDisableCaching(t, w.Header())
	ensure.DeepEqual(t, w.Body.String(), http.StatusText(http.StatusNotFound))
}

func TestServeInvalidData(t *testing.T) {
	h := Handler{Path: "/"}
	w := httptest.NewRecorder()
	r := &http.Request{
		URL: &url.URL{
			Path: "/bar",
		},
	}
	h.ServeHTTP(w, r)
	ensure.DeepEqual(t, w.Code, http.StatusBadRequest)
	ensureDisableCaching(t, w.Header())
	ensure.DeepEqual(t, w.Body.String(), http.StatusText(http.StatusBadRequest))
}

func TestServeLoadError(t *testing.T) {
	h := Handler{
		Path: "/",
		Box: funcBox(func(name string) ([]byte, error) {
			return nil, errors.New("")
		}),
	}
	w := httptest.NewRecorder()
	r := &http.Request{
		URL: &url.URL{
			Path: "/W1sibjEuanMiLCJhY2JkMThkYiJdLFsibjIiLCIzN2I1MWQxOSJdXQ==.js",
		},
	}
	h.ServeHTTP(w, r)
	ensure.DeepEqual(t, w.Code, http.StatusNotFound)
	ensureDisableCaching(t, w.Header())
	ensure.DeepEqual(t, w.Body.String(), http.StatusText(http.StatusNotFound))
}

func TestServeHashMismatch(t *testing.T) {
	h := Handler{
		Path: "/",
		Box: funcBox(func(name string) ([]byte, error) {
			return []byte("foo"), nil
		}),
	}
	w := httptest.NewRecorder()
	r := &http.Request{
		URL: &url.URL{
			Path: "/W1siZm9vIiwiYmFyIl1d",
		},
	}
	h.ServeHTTP(w, r)
	ensure.DeepEqual(t, w.Code, http.StatusNotFound)
	ensureDisableCaching(t, w.Header())
	ensure.DeepEqual(t, w.Body.String(), http.StatusText(http.StatusNotFound))
}

func TestLinkStyleInvalidHREF(t *testing.T) {
	givenErr := errors.New("")
	l := LinkStyle{
		Handler: &Handler{
			Box: funcBox(func(name string) ([]byte, error) {
				return nil, givenErr
			}),
		},
		HREF: []string{"foo"},
	}
	v, err := l.HTML()
	ensure.Nil(t, v)
	ensure.DeepEqual(t, err, givenErr)
}

func TestLinkStyle(t *testing.T) {
	l := LinkStyle{
		Handler: &Handler{
			Box: funcBox(func(name string) ([]byte, error) {
				return []byte("foo"), nil
			}),
		},
		HREF: []string{"foo"},
	}
	v, err := l.HTML()
	ensure.Nil(t, err)
	ensure.DeepEqual(t, v, &h.LinkStyle{
		HREF: "W1siZm9vIiwiYWNiZDE4ZGIiXV0=",
	})
}

func TestScriptInvalidSrc(t *testing.T) {
	givenErr := errors.New("")
	h := Handler{
		Box: funcBox(func(name string) ([]byte, error) {
			return nil, givenErr
		}),
	}
	l := Script{
		Handler: &h,
		Src:     []string{"foo"},
	}
	v, err := l.HTML()
	ensure.Nil(t, v)
	ensure.DeepEqual(t, err, givenErr)
}

func TestScript(t *testing.T) {
	l := Script{
		Handler: &Handler{
			Box: funcBox(func(name string) ([]byte, error) {
				return []byte("foo"), nil
			}),
		},
		Src:   []string{"foo"},
		Async: true,
	}
	v, err := l.HTML()
	ensure.Nil(t, err)
	ensure.DeepEqual(t, v, &h.Script{
		Src:   "W1siZm9vIiwiYWNiZDE4ZGIiXV0=",
		Async: true,
	})
}

func TestImgInvalidSrc(t *testing.T) {
	givenErr := errors.New("")
	l := Img{
		Handler: &Handler{
			Box: funcBox(func(name string) ([]byte, error) {
				return nil, givenErr
			}),
		},
		Src: "foo",
	}
	v, err := l.HTML()
	ensure.Nil(t, v)
	ensure.DeepEqual(t, err, givenErr)
}

func TestImg(t *testing.T) {
	l := Img{
		Handler: &Handler{
			Box: funcBox(func(name string) ([]byte, error) {
				return []byte("foo"), nil
			}),
		},
		Src:   "foo",
		ID:    "a",
		Class: "b",
		Style: "c",
		Alt:   "d",
	}
	v, err := l.HTML()
	ensure.Nil(t, err)
	ensure.DeepEqual(t, v, &h.Img{
		Src:   "W1siZm9vIiwiYWNiZDE4ZGIiXV0=",
		ID:    l.ID,
		Class: l.Class,
		Style: l.Style,
		Alt:   l.Alt,
	})
}
