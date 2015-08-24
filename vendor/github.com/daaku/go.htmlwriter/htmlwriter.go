// Package htmlwriter provides a io.Writer that escapes only five characters:
// <, >, &, ' and ".
package htmlwriter

import (
	"bytes"
	"fmt"
	"io"
)

type writer struct {
	writer io.Writer
}

const escapedChars = "&'<>\"\r"

func (w *writer) Write(s []byte) (int, error) {
	var total int
	i := bytes.IndexAny(s, escapedChars)
	for i != -1 {
		n, err := w.writer.Write(s[:i])
		total += n
		if err != nil {
			return total, err
		}
		var esc string
		switch s[i] {
		case '&':
			esc = "&amp;"
		case '\'':
			// "&#39;" is shorter than "&apos;" and apos was not in HTML until HTML5.
			esc = "&#39;"
		case '<':
			esc = "&lt;"
		case '>':
			esc = "&gt;"
		case '"':
			// "&#34;" is shorter than "&quot;".
			esc = "&#34;"
		case '\r':
			esc = "&#13;"
		default:
			panic("unrecognized escape character")
		}
		s = s[i+1:]

		n, err = fmt.Fprint(w.writer, esc)
		total += n
		if err != nil {
			return total, err
		}
		i = bytes.IndexAny(s, escapedChars)
	}
	n, err := w.writer.Write(s)
	total += n
	return total, err
}

// New creates a new HTML escaping writer that will write to the given writer.
func New(w io.Writer) io.Writer {
	return &writer{writer: w}
}
