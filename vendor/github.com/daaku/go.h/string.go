package h

import (
	"html"
)

type String string

func (s String) HTML() (HTML, error) {
	return Unsafe(html.EscapeString(string(s))), nil
}
