package view_test

import (
	"github.com/nshah/go.h"
	"github.com/nshah/rell/view"
	"testing"
)

func TestBasePage(t *testing.T) {
	page := &view.Page{
		Title: "the title &",
		Inner: h.String("hello world"),
		Class: "main",
	}

	str, err := h.Render(page)
	t.Errorf("%s %s\n", str, err)
}
