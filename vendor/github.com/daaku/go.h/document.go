package h

import "golang.org/x/net/context"

type XMLNS map[string]string

func (ns XMLNS) Attributes() Attributes {
	attrs := Attributes{}
	for key, val := range ns {
		attrs["xmlns:"+key] = val
	}
	return attrs
}

type Document struct {
	XMLNS XMLNS
	Inner HTML
	ID    string
	Lang  string
}

func (d *Document) HTML(ctx context.Context) (HTML, error) {
	attrs := d.XMLNS.Attributes()
	if d.ID != "" {
		attrs["id"] = d.ID
	}
	if d.Lang != "" {
		attrs["lang"] = "en"
	}

	return &Frag{
		Unsafe("<!doctype html>"),
		&Node{
			Tag:        "html",
			Attributes: attrs,
			Inner:      d.Inner,
		},
	}, nil
}
