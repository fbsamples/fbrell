package examples

import (
	"github.com/nshah/go.h"
	"testing"
)

func TestRenderIndex(t *testing.T) {
	page := NewDB("").RenderIndex()
	str, err := h.Render(page)
	t.Errorf("%s %s\n", str, err)
}
