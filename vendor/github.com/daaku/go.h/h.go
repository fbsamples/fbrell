// Package provides a psuedo-DOM style approach to generating HTML markup.
// **Unstable API. Work in progress.**
package h

import (
	"bytes"
	"fmt"
	"io"
	"log"

	"golang.org/x/net/context"
)

type HTML interface {
	HTML(context.Context) (HTML, error)
}

type Primitive interface {
	Write(context.Context, io.Writer) (int, error)
}

// Write HTML into a writer.
func Write(ctx context.Context, w io.Writer, h HTML) (int, error) {
	var err error
	for {
		switch t := h.(type) {
		case nil:
			return 0, nil
		case Primitive:
			return t.Write(ctx, w)
		case HTML:
			h, err = h.HTML(ctx)
			if err != nil {
				return 0, err
			}
		default:
			return 0, fmt.Errorf("Value %+v of unknown type %T", h, h)
		}
	}
}

// Render HTML as a string.
func Render(ctx context.Context, h HTML) (string, error) {
	buffer := bytes.NewBufferString("")
	_, err := Write(ctx, buffer, h)
	return buffer.String(), err
}

// Compile static HTML into HTML. Will panic if there are errors.
func Compile(ctx context.Context, h HTML) HTML {
	m, err := Render(ctx, h)
	if err != nil {
		log.Fatalf("Failed to Compile HTML %v with error %s", h, err)
	}
	return Unsafe(m)
}
