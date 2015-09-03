// Package ga provides go.h compatible async loading for Google Analytics.
package ga

import (
	"errors"
	"fmt"

	"golang.org/x/net/context"

	"github.com/daaku/go.h"
)

var ErrMissingID = errors.New("GoogleAnalyics requires an ID.")

// Loadable for a Page Track event using Google Analytics.
type Track struct {
	ID string
}

func (g *Track) HTML(ctx context.Context) (h.HTML, error) {
	if g.ID == "" {
		return nil, ErrMissingID
	}
	return &h.Frag{
		&h.Script{
			Inner: h.Unsafe(fmt.Sprintf(
				`var _gaq = _gaq || [];`+
					`_gaq.push(['_setAccount', '%s']);`+
					`_gaq.push(['_trackPageview']);`, g.ID)),
		},
		&h.Script{
			Src:   "https://ssl.google-analytics.com/ga.js",
			Async: true,
		},
	}, nil
}
