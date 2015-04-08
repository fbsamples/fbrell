package htmlwriter_test

import (
	"bytes"
	"html"
	"testing"

	"github.com/daaku/rell/internal/github.com/daaku/go.htmlwriter"
)

func TestSimple(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	w := htmlwriter.New(&out)
	const original = "hello & world"
	n, err := w.Write([]byte(original))
	if err != nil {
		t.Fatal(err)
	}
	const l = 17
	if n != l {
		t.Fatalf("expected %d but got %d", l, n)
	}
	expected := html.EscapeString(original)
	if out.String() != expected {
		t.Fatalf("expected %s but got %s", expected, out)
	}
}
