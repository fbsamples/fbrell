// Package ga provides go.h compatible async loading for Google Analytics.
package ga

import (
	"errors"
	"fmt"

	"golang.org/x/net/context"

	"github.com/daaku/go.h"
)

var errMissingAccount = errors.New("ga: missing required Account field")

// Track adds the basic Google Analytics Page Tracking. More here:
// https://developers.google.com/analytics/devguides/collection/gajs/
type Track struct {
	Account string
}

// HTML renders the relevant <script> tags.
func (g *Track) HTML(ctx context.Context) (h.HTML, error) {
	if g.Account == "" {
		return nil, errMissingAccount
	}
	return h.Frag{
		&h.Script{
			Inner: h.Unsafe(fmt.Sprintf(
				`var _gaq = _gaq || [];`+
					`_gaq.push(['_setAccount', '%s']);`+
					`_gaq.push(['_trackPageview']);`, g.Account)),
		},
		&h.Script{
			Src:   "https://ssl.google-analytics.com/ga.js",
			Async: true,
		},
	}, nil
}
