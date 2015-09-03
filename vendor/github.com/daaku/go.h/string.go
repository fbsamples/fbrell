package h

import (
	"html"

	"golang.org/x/net/context"
)

type String string

func (s String) HTML(ctx context.Context) (HTML, error) {
	return Unsafe(html.EscapeString(string(s))), nil
}
