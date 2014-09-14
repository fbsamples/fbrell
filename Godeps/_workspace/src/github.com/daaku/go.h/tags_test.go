package h_test

import (
	"github.com/daaku/go.h"
	"testing"
)

func TestEmptyBody(t *testing.T) {
	t.Parallel()
	body := &h.Body{}
	assertRender(t, body, `<body></body>`)
}

func TestEmptyHead(t *testing.T) {
	t.Parallel()
	head := &h.Head{}
	assertRender(t, head, `<head></head>`)
}

func TestEmptyH1(t *testing.T) {
	t.Parallel()
	h1 := &h.H1{}
	assertRender(t, h1, `<h1></h1>`)
}

func TestEmptyLink(t *testing.T) {
	t.Parallel()
	link := &h.Link{}
	assertRender(t, link, `<link>`)
}

func TestEmptyLi(t *testing.T) {
	t.Parallel()
	li := &h.Li{}
	assertRender(t, li, `<li></li>`)
}

func TestEmptyMeta(t *testing.T) {
	t.Parallel()
	meta := &h.Meta{}
	assertRender(t, meta, `<meta>`)
}

func TestEmptySpan(t *testing.T) {
	t.Parallel()
	span := &h.Span{}
	assertRender(t, span, `<span></span>`)
}

func TestEmptyTitle(t *testing.T) {
	t.Parallel()
	title := &h.Title{h.String("")}
	assertRender(t, title, `<title></title>`)
}
