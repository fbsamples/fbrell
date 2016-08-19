// Package static provides go.h compatible hashed static assets. This allows
// for providing long lived cache headers for resources which change URLs as
// their content changes.
package static

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/daaku/go.h"
)

const (
	maxAge  = time.Hour * 24 * 365 * 10
	hashLen = 8
)

var (
	errZeroNames          = errors.New("static: zero names given")
	errNoHandlerInContext = errors.New("static: no handler in context")
	cacheControl          = fmt.Sprintf("public, max-age=%d", int(maxAge.Seconds()))
)

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

// we drop base64 padding in our URLs
func dropPadding(s string) string {
	return strings.TrimRight(s, "=")
}

// so we only allocate these strings once
var paddings = [...]string{
	"=",
	"==",
	"===",
}

func addPadding(s string) string {
	if l := len(s) % 4; l > 0 {
		return s + paddings[3-l]
	}
	return s
}

type file struct {
	Name    string
	Content []byte
	Hash    string
}

func encode(files []file) (string, error) {
	parts := make([][2]string, 0, len(files))
	for _, f := range files {
		parts = append(parts, [2]string{f.Name, f.Hash})
	}

	b, err := json.Marshal(parts)
	if err != nil {
		return "", errors.New("static: could not encode URL")
	}

	return dropPadding(base64.URLEncoding.EncodeToString(b)), nil
}

func decode(value string) ([]file, error) {
	decoded, err := base64.URLEncoding.DecodeString(addPadding(value))
	if err != nil {
		return nil, errInvalidURL(value)
	}

	var parts [][2]string
	if err := json.NewDecoder(bytes.NewReader(decoded)).Decode(&parts); err != nil {
		return nil, errInvalidURL(value)
	}

	files := make([]file, 0, len(parts))
	for _, part := range parts {
		if len(part) != 2 || part[0] == "" || part[1] == "" {
			return nil, errInvalidURL(value)
		}
		files = append(files, file{
			Name: part[0],
			Hash: part[1],
		})
	}

	return files, nil
}

// Box is where the files are loaded from. You'll probably want to use
// https://github.com/GeertJohan/go.rice or FileSystemBox.
type Box interface {
	Bytes(name string) ([]byte, error)
}

type fileSystemBox struct {
	fs http.FileSystem
}

func (b *fileSystemBox) Bytes(name string) ([]byte, error) {
	f, err := b.fs.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

// FileSystemBox returns a Box from a http.FileSystem.
func FileSystemBox(fs http.FileSystem) Box {
	return &fileSystemBox{fs: fs}
}

// Handler serves and provides URLs for static resources.
type Handler struct {
	Path string // Path at which Handler is configured.
	Box  Box    // Box of files to serve.

	mu    sync.RWMutex
	files map[string]file
}

func (h *Handler) load(name string) (file, error) {
	// fast path
	h.mu.RLock()
	f, found := h.files[name]
	h.mu.RUnlock()

	if found {
		return f, nil
	}

	// slow path
	h.mu.Lock()
	defer h.mu.Unlock()

	// check again in case someone else populated it
	f, found = h.files[name]
	if found {
		return f, nil
	}

	contents, err := h.Box.Bytes(name)
	if err != nil {
		return file{}, err
	}

	hash := fmt.Sprintf("%x", md5.Sum(contents))
	f = file{
		Name:    name,
		Content: contents,
		Hash:    hash[:hashLen],
	}
	if h.files == nil {
		h.files = make(map[string]file)
	}
	h.files[name] = f

	return f, nil
}

// URL returns a hashed URL for all the given component names. It uses the
// extension of the first file as the extension for the generated URL.
func (h *Handler) URL(names ...string) (string, error) {
	if len(names) == 0 {
		return "", errZeroNames
	}

	files := make([]file, 0, len(names))
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
	header.Set("Cache-Control", cacheControl)
	header.Set("Content-Length", strconv.Itoa(contentLength))
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
	HREF []string
}

// HTML returns the <link> tag with the appropriate attributes.
func (l *LinkStyle) HTML(ctx context.Context) (h.HTML, error) {
	url, err := URL(ctx, l.HREF...)
	if err != nil {
		return nil, err
	}
	return &h.LinkStyle{HREF: url}, nil
}

// Script provides a h.Script where the Srcs are combined and served using the
// specified Handler.
type Script struct {
	Src   []string
	Async bool
}

// HTML returns the <script> tag with the appropriate attributes.
func (l *Script) HTML(ctx context.Context) (h.HTML, error) {
	url, err := URL(ctx, l.Src...)
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
	ID    string
	Class string
	Style string
	Src   string
	Alt   string
}

// HTML returns the <img> tag with the appropriate attributes.
func (i *Img) HTML(ctx context.Context) (h.HTML, error) {
	src, err := URL(ctx, i.Src)
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

// Favicon provides a h.Link for a favicon.
type Favicon struct {
	HREF string
}

// HTML returns the <script> tag with the appropriate attributes.
func (l *Favicon) HTML(ctx context.Context) (h.HTML, error) {
	url, err := URL(ctx, l.HREF)
	if err != nil {
		return nil, err
	}
	return &h.Link{
		Rel:  "shortcut icon",
		HREF: url,
	}, nil
}

// Input renders a HTML <input> tag with the Src URL transformed.
type Input struct {
	ID          string
	Class       string
	Name        string
	Style       string
	Type        string
	Value       string
	Src         string
	Placeholder string
	Checked     bool
	Multiple    bool
	Data        map[string]interface{}
	Inner       h.HTML
}

// HTML renders the content.
func (i *Input) HTML(ctx context.Context) (h.HTML, error) {
	src, err := URL(ctx, i.Src)
	if err != nil {
		return nil, err
	}
	return &h.Input{
		ID:          i.ID,
		Class:       i.Class,
		Name:        i.Name,
		Style:       i.Style,
		Type:        i.Type,
		Value:       i.Value,
		Src:         src,
		Placeholder: i.Placeholder,
		Checked:     i.Checked,
		Multiple:    i.Multiple,
		Data:        i.Data,
		Inner:       i.Inner,
	}, nil
}

type ctxKey int

const handlerCtxKey ctxKey = 0

func NewContext(ctx context.Context, h *Handler) context.Context {
	return context.WithValue(ctx, handlerCtxKey, h)
}

func FromContext(ctx context.Context) *Handler {
	h, _ := ctx.Value(handlerCtxKey).(*Handler)
	return h
}

func URL(ctx context.Context, names ...string) (string, error) {
	h := FromContext(ctx)
	if h == nil {
		return "", errNoHandlerInContext
	}
	return h.URL(names...)
}
