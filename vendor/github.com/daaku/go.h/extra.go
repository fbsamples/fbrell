package h

import (
	"errors"
	"net/url"

	"golang.org/x/net/context"
)

var errLinkStyleMissingHREF = errors.New("h: LinkStyle missing HREF")

// HiddenInputs renders some HTML for the given url.Values. It does so inside a
// div with display: none.
func HiddenInputs(values url.Values) HTML {
	var frag Frag
	for key, list := range values {
		for _, each := range list {
			frag = append(frag, &Input{Name: key, Value: each})
		}
	}
	return &Div{Style: "display:none", Inner: frag}
}

var _ HTML = (*LinkStyle)(nil)

// LinkStyle provides the common CSS Style Sheet tag.
type LinkStyle struct {
	HREF string
}

// HTML renders the tag.
func (l *LinkStyle) HTML(ctx context.Context) (HTML, error) {
	if l.HREF == "" {
		return nil, errLinkStyleMissingHREF
	}
	return &Link{
		Type: "text/css",
		Rel:  "stylesheet",
		HREF: l.HREF,
	}, nil
}
