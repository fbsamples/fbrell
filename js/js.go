// Package js provides the Rell JavaScript resources.
package js

import (
	"encoding/json"
	"fmt"

	"github.com/daaku/go.h"
	"github.com/daaku/rell/context"
	"github.com/daaku/rell/examples"
)

// Represents configuration for initializing the rell module. Sets up a couple
// of globals.
type Init struct {
	Context *context.Context
	Example *examples.Example
}

func (i *Init) HTML() (h.HTML, error) {
	encodedContext, err := json.Marshal(i.Context)
	if err != nil {
		return nil, fmt.Errorf("Failed to json.Marshal context: %s", err)
	}
	encodedExample, err := json.Marshal(i.Example)
	if err != nil {
		return nil, fmt.Errorf("Failed to json.Marshal example: %s", err)
	}
	return &h.Frag{
		&h.Script{
			Src:   i.Context.SdkURL(),
			Async: true,
		},
		&h.Script{
			Inner: &h.Frag{
				h.Unsafe("window.rellConfig="),
				h.UnsafeBytes(encodedContext),
				h.Unsafe(";window.rellExample="),
				h.UnsafeBytes(encodedExample),
			},
		},
	}, nil
}
