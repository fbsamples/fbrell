package h

import (
	"fmt"
	"io"

	"golang.org/x/net/context"
)

type Frag []HTML

func (f *Frag) HTML(ctx context.Context) (HTML, error) {
	return f, fmt.Errorf("Frag.HTML called for %s", f)
}

func (f *Frag) Append(h HTML) {
	*f = append(*f, h)
}

func (f *Frag) Write(ctx context.Context, w io.Writer) (int, error) {
	written := 0
	for _, e := range *f {
		i, err := Write(ctx, w, e)
		written += i
		if err != nil {
			return written, err
		}
	}
	return written, nil
}
