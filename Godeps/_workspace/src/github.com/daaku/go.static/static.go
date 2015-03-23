// Package static provides go.h compatible hashed static asset
// URIs. This allows for providing long lived cache headers for
// resources which change URLs as their content changes.
package static

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/daaku/rell/Godeps/_workspace/src/github.com/GeertJohan/go.rice"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.h"
)

const errHandlerRequired = "go.static: a handler is required for static HTML: %+v"

type cacheEntry struct {
	Content []byte
	ModTime time.Time
}

type Handler struct {
	sync.RWMutex
	HttpPath    string        // prefix path for static files
	MaxAge      time.Duration // max-age for HTTP headers
	MemoryCache bool          // enable in memory cache
	Box         *rice.Box
	cache       map[string]cacheEntry
}

func notFound(w http.ResponseWriter, r *http.Request) {
	header := w.Header()
	header.Set("Cache-Control", "no-cache")
	header.Set("Pragma", "no-cache")
	w.WriteHeader(404)
	w.Write([]byte("static resource not found!"))
}

func joinBasenames(names []string) string {
	basenames := make([]string, len(names))
	for i, name := range names {
		basenames[i] = filepath.Base(name)
	}
	return strings.Join(basenames, "-")
}

// Get a hashed URL for a single file.
func (h *Handler) URL(name string) (string, error) {
	return h.CombinedURL([]string{name})
}

// Get a hashed combined URL for all named files.
func (h *Handler) CombinedURL(names []string) (string, error) {
	hash := md5.New()
	var ce cacheEntry
	for _, name := range names {
		f, err := h.Box.HTTPBox().Open(name)
		if err != nil {
			return "", err
		}
		defer f.Close()

		stat, err := f.Stat()
		if err != nil {
			return "", err
		}
		modTime := stat.ModTime()
		if ce.ModTime.Before(modTime) {
			ce.ModTime = modTime
		}

		content, err := ioutil.ReadAll(f)
		if err != nil {
			return "", err
		}
		ce.Content = append(ce.Content, content...)
		_, err = hash.Write(content)
		if err != nil {
			return "", err
		}
	}
	hex := fmt.Sprintf("%x", hash.Sum(nil))
	hexS := hex[:10]
	url := path.Join(h.HttpPath, hexS, joinBasenames(names))
	h.Lock()
	defer h.Unlock()
	if h.cache == nil {
		h.cache = make(map[string]cacheEntry)
	}
	h.cache[hexS] = ce
	return url, nil
}

// Serves the static resource.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if !strings.HasPrefix(path, h.HttpPath) {
		notFound(w, r)
		return
	}
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		notFound(w, r)
		return
	}

	h.RLock()
	defer h.RUnlock()
	if h.cache == nil {
		notFound(w, r)
		return
	}
	ce, ok := h.cache[parts[2]]
	if !ok {
		notFound(w, r)
		return
	}

	header := w.Header()
	header.Set(
		"Cache-Control",
		fmt.Sprintf("public, max-age=%d", int(h.MaxAge.Seconds())))
	http.ServeContent(w, r, path, ce.ModTime, bytes.NewReader(ce.Content))
}

type LinkStyle struct {
	HREF    []string
	Handler *Handler
	cache   h.HTML
}

func (l *LinkStyle) HTML() (h.HTML, error) {
	if l.Handler == nil {
		return nil, fmt.Errorf(errHandlerRequired, l)
	}
	if !l.Handler.MemoryCache || l.cache == nil {
		url, err := l.Handler.CombinedURL(l.HREF)
		if err != nil {
			return nil, err
		}
		l.cache = &h.LinkStyle{HREF: url}
	}
	return l.cache, nil
}

type Script struct {
	Src     []string
	Async   bool
	Handler *Handler
	cache   h.HTML
}

func (l *Script) HTML() (h.HTML, error) {
	if l.Handler == nil {
		return nil, fmt.Errorf(errHandlerRequired, l)
	}
	if !l.Handler.MemoryCache || l.cache == nil {
		url, err := l.Handler.CombinedURL(l.Src)
		if err != nil {
			return nil, err
		}
		l.cache = &h.Script{
			Src:   url,
			Async: l.Async,
		}
	}
	return l.cache, nil
}

type Img struct {
	ID      string
	Class   string
	Style   string
	Src     string
	Alt     string
	Handler *Handler
	cache   h.HTML
}

func (i *Img) HTML() (h.HTML, error) {
	if i.Handler == nil {
		return nil, fmt.Errorf(errHandlerRequired, i)
	}
	if !i.Handler.MemoryCache || i.cache == nil {
		src, err := i.Handler.URL(i.Src)
		if err != nil {
			return nil, err
		}
		i.cache = &h.Node{
			Tag:         "img",
			SelfClosing: true,
			Attributes: h.Attributes{
				"id":    i.ID,
				"class": i.Class,
				"style": i.Style,
				"src":   src,
				"alt":   i.Alt,
			},
		}
	}
	return i.cache, nil
}
