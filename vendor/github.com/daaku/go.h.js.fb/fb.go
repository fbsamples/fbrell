// Package fb provides go.h compatible async loading for the Facebook JS SDK.
package fb

import (
	"encoding/json"

	"golang.org/x/net/context"

	"github.com/daaku/go.h"
)

// Init is akin to the FB init call for the Facebook JS SDK.
type Init struct {
	URL                  string `json:"-"`
	AppID                uint64 `json:"appId"`
	Version              string `json:"version"`
	Cookie               bool   `json:"cookie,omitempty"`
	Status               bool   `json:"status,omitempty"`
	XFBML                bool   `json:"xfbml,omitempty"`
	FrictionlessRequests bool   `json:"frictionlessRequests,omitempty"`
}

const defaultURL = "//connect.facebook.net/en_US/all.js"

// HTML returns the pair of <script> tags that load and render the SDK.
func (i *Init) HTML(ctx context.Context) (h.HTML, error) {
	url := i.URL
	if url == "" {
		url = defaultURL
	}

	encoded, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}
	return h.Frag{
		&h.Script{
			Src:   url,
			Async: true,
		},
		&h.Script{
			Inner: h.Frag{
				h.Unsafe("window.fbAsyncInit=function(){FB.init("),
				h.UnsafeBytes(encoded),
				h.Unsafe(")"),
			},
		},
	}, nil
}
