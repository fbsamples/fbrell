package h

import (
	"fmt"
	"io"

	"golang.org/x/net/context"
)

type UnsafeBytes []byte

func (u UnsafeBytes) HTML(ctx context.Context) (HTML, error) {
	return u, fmt.Errorf("UnsafeBytes.HTML called for %s", u)
}

func (u UnsafeBytes) Write(ctx context.Context, w io.Writer) (int, error) {
	return fmt.Fprintf(w, "%s", u)
}
