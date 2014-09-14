package h_test

import (
	"github.com/daaku/go.h"
	"testing"
)

func assertRender(t *testing.T, html h.HTML, expected string) {
	actual, err := h.Render(html)
	if err != nil {
		t.Fatalf("Failed to render document %v with error %s", html, err)
	}
	if actual != expected {
		t.Fatalf("Did not find expected:\n%s\ninstead found:\n%s", expected, actual)
	}
}
