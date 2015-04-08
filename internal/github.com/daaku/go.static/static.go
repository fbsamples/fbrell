// Package static provides go.h compatible hashed static assets. This allows
// for providing long lived cache headers for resources which change URLs as
// their content changes.
package static

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/daaku/rell/internal/github.com/daaku/go.h"
)

const maxAge = time.Hour * 24 * 365 * 10

var errZeroNames = errors.New("static: zero names given")

func disableCaching(w http.ResponseWriter) {
	header := w.Header()
	header.Set("Cache-Control", "no-cache")
	header.Set("Pragma", "no-cache")
}

func notFound(w http.ResponseWriter) {
	disableCaching(w)
	w.WriteHeader(http.StatusNotFound)
	io.WriteString(w, http.StatusText(http.StatusNotFound))
}

func badRequest(w http.ResponseWriter) {
	disableCaching(w)
	w.WriteHeader(http.StatusBadRequest)
	io.WriteString(w, http.StatusText(http.StatusBadRequest))
}

type errInvalidURL string

func (e errInvalidURL) Error() string {
	return fmt.Sprintf("static: invalid URL %q", string(e))
}

type file struct {
	Name    string
	Content []byte
	Hash    string
}

func encode(files []*file) (string, error) {
	var parts [][]string
	for _, f := range files {
		parts = append(parts, []string{f.Name, f.Hash})
	}

	b, err := json.Marshal(parts)
	if err != nil {
		return "", errors.New("static: could not encode URL")
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

func decode(value string) ([]*file, error) {
	decoded, err := base64.URLEncoding.DecodeString(value)
	if err != nil {
		return nil, errInvalidURL(value)
	}

	var parts [][]string
	if err := json.NewDecoder(bytes.NewReader(decoded)).Decode(&parts); err != nil {
		return nil, errInvalidURL(value)
	}

	var files []*file
	for _, part := range parts {
		if len(part) != 2 {
			return nil, errInvalidURL(value)
		}
		files = append(files, &file{
			Name: part[0],
			Hash: part[1],
		})
	}

	return files, nil
}

// Box is where the files are loaded from. Practically you'll probably want to
// use https://github.com/GeertJohan/go.rice.
type Box interface {
	Bytes(name string) ([]byte, error)
}

// Handler serves and provides URLs for static resources.
type Handler struct {
	Path string // Path at which Handler is configured.
	Box  Box    // Box of files to serve.

	mu    sync.RWMutex
	files map[string]*file
}

func (h *Handler) load(name string) (*file, error) {
	// fast path
	h.mu.RLock()
	f := h.files[name]
	h.mu.RUnlock()

	if f != nil {
		return f, nil
	}

	// slow path
	h.mu.Lock()
	defer h.mu.Unlock()

	// check again in case someone else populated it
	f = h.files[name]
	if f != nil {
		return f, nil
	}

	contents, err := h.Box.Bytes(name)
	if err != nil {
		return nil, err
	}

	hash := fmt.Sprintf("%x", md5.Sum(contents))
	f = &file{
		Name:    name,
		Content: contents,
		Hash:    hash[:8],
	}
	if h.files == nil {
		h.files = make(map[string]*file)
	}
	h.files[name] = f

	return f, nil
}

// URL returns a hashed URL for all the given component names.
func (h *Handler) URL(names ...string) (string, error) {
	if len(names) == 0 {
		return "", errZeroNames
	}

	var files []*file
	for _, name := range names {
		f, err := h.load(name)
		if err != nil {
			return "", err
		}
		files = append(files, f)
	}

	value, err := encode(files)
	if err != nil {
		return "", err
	}

	if ext := filepath.Ext(names[0]); ext != "" {
		value = value + ext
	}

	return path.Join(h.Path, value), nil
}

// ServeHTTP handles requests for hashed URLs.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if !strings.HasPrefix(path, h.Path) {
		notFound(w)
		return
	}

	contentType := ""
	encoded := path[len(h.Path):]
	if ext := filepath.Ext(encoded); ext != "" {
		encoded = encoded[:len(encoded)-len(ext)]
		contentType = mime.TypeByExtension(ext)
	}

	files, err := decode(encoded)
	if err != nil {
		badRequest(w)
		return
	}

	// fill in the contents and calculate the length
	var contentLength int
	for i, f := range files {
		loaded, err := h.load(f.Name)
		if err != nil {
			notFound(w)
			return
		}
		if loaded.Hash != f.Hash {
			notFound(w)
			return
		}
		contentLength += len(loaded.Content)
		files[i] = loaded
	}

	header := w.Header()
	header.Set(
		"Cache-Control",
		fmt.Sprintf("public, max-age=%d", int(maxAge.Seconds())))
	header.Set("Content-Length", fmt.Sprint(contentLength))
	if contentType != "" {
		header.Set("Content-Type", contentType)
	}

	for _, f := range files {
		w.Write(f.Content)
	}
}

// LinkStyle provides a h.LinkStyle where the HREFs are combined and served
// using the specified Handler.
type LinkStyle struct {
	Handler *Handler
	HREF    []string
}

// HTML returns the <link> tag with the appropriate attributes.
func (l *LinkStyle) HTML() (h.HTML, error) {
	url, err := l.Handler.URL(l.HREF...)
	if err != nil {
		return nil, err
	}
	return &h.LinkStyle{HREF: url}, nil
}

// Script provides a h.Script where the Srcs are combined and served using the
// specified Handler.
type Script struct {
	Handler *Handler
	Src     []string
	Async   bool
}

// HTML returns the <script> tag with the appropriate attributes.
func (l *Script) HTML() (h.HTML, error) {
	url, err := l.Handler.URL(l.Src...)
	if err != nil {
		return nil, err
	}
	return &h.Script{
		Src:   url,
		Async: l.Async,
	}, nil
}

// Img provides a h.Img where the src is served using the specified Handler.
type Img struct {
	Handler *Handler
	ID      string
	Class   string
	Style   string
	Src     string
	Alt     string
}

// HTML returns the <img> tag with the appropriate attributes.
func (i *Img) HTML() (h.HTML, error) {
	src, err := i.Handler.URL(i.Src)
	if err != nil {
		return nil, err
	}
	return &h.Img{
		ID:    i.ID,
		Class: i.Class,
		Style: i.Style,
		Src:   src,
		Alt:   i.Alt,
	}, nil
}
