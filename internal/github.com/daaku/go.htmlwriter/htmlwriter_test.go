package htmlwriter_test

import (
	"bytes"
	"html"
	"testing"

	"github.com/daaku/rell/internal/github.com/daaku/go.htmlwriter"
	"github.com/daaku/rell/internal/github.com/facebookgo/ensure"
)

func TestSimple(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	w := htmlwriter.New(&out)
	const original = "hello & world"
	n, err := w.Write([]byte(original))
	ensure.Nil(t, err)
	ensure.DeepEqual(t, n, 17)
	ensure.DeepEqual(t, out.String(), html.EscapeString(original))
}
