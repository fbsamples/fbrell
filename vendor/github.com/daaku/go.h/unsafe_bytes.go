package h

import (
	"fmt"
	"io"
)

type UnsafeBytes []byte

func (u UnsafeBytes) HTML() (HTML, error) {
	return u, fmt.Errorf("UnsafeBytes.HTML called for %s", u)
}

func (u UnsafeBytes) Write(w io.Writer) (int, error) {
	return fmt.Fprintf(w, "%s", u)
}
