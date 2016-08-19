// Package ga provides go.h compatible async loading for Google Analytics.
package ga

import (
	"context"
	"errors"

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
	return &h.Script{
		Inner: h.Frag{
			h.Unsafe(
				`(function(i,s,o,g,r,a,m){i['GoogleAnalyticsObject']=r;i[r]=i[r]||function(){` +
					`(i[r].q=i[r].q||[]).push(arguments)},i[r].l=1*new Date();a=s.createElement(o),` +
					`m=s.getElementsByTagName(o)[0];a.async=1;a.src=g;m.parentNode.insertBefore(a,m)` +
					`})(window,document,'script','//www.google-analytics.com/analytics.js','ga');` +
					`ga('create','`),
			h.String(g.Account),
			h.Unsafe(`','auto');ga('send','pageview');`),
		},
	}, nil
}
