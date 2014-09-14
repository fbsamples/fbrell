package h_test

import (
	"github.com/daaku/go.h"
	"testing"
)

func TestEmptyDocument(t *testing.T) {
	t.Parallel()
	doc := &h.Document{}
	assertRender(t, doc, `<!doctype html><html></html>`)
}

func TestFacebookXMLNS(t *testing.T) {
	t.Parallel()
	doc := &h.Document{
		XMLNS: h.XMLNS{
			"fb": "http://ogp.me/ns/fb#",
		},
	}
	assertRender(t, doc,
		`<!doctype html><html xmlns:fb="http://ogp.me/ns/fb#"></html>`)
}
