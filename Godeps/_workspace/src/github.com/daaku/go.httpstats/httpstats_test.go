package httpstats_test

import (
	"github.com/daaku/go.httpstats"
	"testing"
)

func TestCreateHandler(t *testing.T) {
	httpstats.NewHandler("web", nil)
}
