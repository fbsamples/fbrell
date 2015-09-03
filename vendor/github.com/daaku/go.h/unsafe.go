package h

import (
	"fmt"
	"io"

	"golang.org/x/net/context"
)

type Unsafe string

func (u Unsafe) HTML(ctx context.Context) (HTML, error) {
	return u, fmt.Errorf("Unsafe.HTML called for %s", u)
}

func (u Unsafe) Write(ctx context.Context, w io.Writer) (int, error) {
	return fmt.Fprint(w, u)
}
