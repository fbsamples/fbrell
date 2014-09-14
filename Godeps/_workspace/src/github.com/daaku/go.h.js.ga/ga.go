// Package ga provides go.h compatible async loading for
// Google Analytics.
package ga

import (
	"errors"
	"fmt"
	"github.com/daaku/go.h"
	"log"
)

var ErrMissingID = errors.New("GoogleAnalyics requires an ID.")

// Loadable for a Page Track event using Google Analytics.
type Track struct {
	ID string
}

func (g *Track) URLs() []string {
	return []string{"https://ssl.google-analytics.com/ga.js"}
}

func (g *Track) Script() string {
	if g.ID == "" {
		log.Fatal("GoogleAnalyics requires an ID.")
	}
	return fmt.Sprintf(
		`try{`+
			`var pageTracker=_gat._getTracker("%s");`+
			`pageTracker._trackPageview();`+
			`}catch(e){}`, g.ID)
}

func (g *Track) HTML() (h.HTML, error) {
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
			Src:   g.URLs()[0],
			Async: true,
		},
	}, nil
}
