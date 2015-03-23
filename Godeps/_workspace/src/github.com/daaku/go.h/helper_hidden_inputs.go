package h

import (
	"net/url"
)

func HiddenInputs(values url.Values) HTML {
	frag := &Frag{}
	for key, list := range values {
		for _, each := range list {
			frag.Append(&Input{Name: key, Value: each})
		}
	}
	return &Div{Style: "display:none", Inner: frag}
}
