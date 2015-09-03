package h

import (
	"errors"

	"golang.org/x/net/context"
)

type LinkStyle struct {
	HREF string
}

func (l *LinkStyle) HTML(ctx context.Context) (HTML, error) {
	if l.HREF == "" {
		return nil, errors.New("Missing HREF in LinkStyle.")
	}
	return &Link{
		Type: "text/css",
		Rel:  "stylesheet",
		HREF: l.HREF,
	}, nil
}
