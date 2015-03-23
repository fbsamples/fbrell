package static_test

import (
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.static"
	"testing"
)

func TestSanity(t *testing.T) {
	h := static.Handler{}
	h.URL("a")
}
