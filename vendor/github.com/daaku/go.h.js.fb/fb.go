// Package fb provides go.h compatible async loading for the Facebook JS SDK.
package fb

import (
	"encoding/json"
	"fmt"

	"golang.org/x/net/context"

	"github.com/daaku/go.h"
)

// Represents an async load and FB.init call for the Facebook JS SDK.
type Init struct {
	URL   string `json:"-"`
	AppID uint64 `json:"appId"`
}

const defaultURL = "//connect.facebook.net/en_US/all.js"

func (i *Init) HTML(ctx context.Context) (h.HTML, error) {
	url := i.URL
	if url == "" {
		url = defaultURL
	}

	encoded, err := json.Marshal(i)
	if err != nil {
		return nil, fmt.Errorf("Failed to json.Marshal sdk.Init %+v with error %s", i, err)
	}
	return &h.Frag{
		&h.Script{
			Src:   url,
			Async: true,
		},
		&h.Script{
			Inner: &h.Frag{
				h.Unsafe("window.fbAsyncInit=function(){FB.init("),
				h.UnsafeBytes(encoded),
				h.Unsafe(")"),
			},
		},
	}, nil
}
