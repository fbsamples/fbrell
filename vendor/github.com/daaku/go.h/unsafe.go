package h

import (
	"fmt"
	"io"
)

type Unsafe string

func (u Unsafe) HTML() (HTML, error) {
	return u, fmt.Errorf("Unsafe.HTML called for %s", u)
}

func (u Unsafe) Write(w io.Writer) (int, error) {
	return fmt.Fprint(w, u)
}
