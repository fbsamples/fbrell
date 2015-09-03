package h

import (
	"fmt"
	"io"

	"golang.org/x/net/context"
)

type Node struct {
	Tag         string
	Attributes  Attributes
	Inner       HTML
	SelfClosing bool
}

func (n *Node) HTML(ctx context.Context) (HTML, error) {
	return n, fmt.Errorf("Called HTML for Node: %+v", n)
}

// Write the generated markup for a Node.
func (n *Node) Write(ctx context.Context, w io.Writer) (int, error) {
	written := 0
	i := 0
	var err error

	i, err = fmt.Fprint(w, "<", n.Tag)
	written += i
	if err != nil {
		return written, err
	}

	i, err = n.Attributes.Write(w, "")
	written += i
	if err != nil {
		return written, err
	}

	i, err = fmt.Fprint(w, ">")
	written += i
	if err != nil {
		return written, err
	}

	i, err = Write(ctx, w, n.Inner)
	written += i
	if err != nil {
		return written, err
	}

	if !n.SelfClosing {
		i, err = fmt.Fprint(w, "</", n.Tag, ">")
		written += i
		if err != nil {
			return written, err
		}
	}

	return written, nil
}
