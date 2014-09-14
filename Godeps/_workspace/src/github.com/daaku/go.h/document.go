package h

type XMLNS map[string]string

type Document struct {
	XMLNS XMLNS
	Inner HTML
}

func (ns XMLNS) Attributes() Attributes {
	attrs := Attributes{}
	for key, val := range ns {
		attrs["xmlns:"+key] = val
	}
	return attrs
}

func (d *Document) HTML() (HTML, error) {
	return &Frag{
		Unsafe("<!doctype html>"),
		&Node{
			Tag:        "html",
			Attributes: d.XMLNS.Attributes(),
			Inner:      d.Inner,
		},
	}, nil
}
