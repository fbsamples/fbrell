// Package loader provides async script loading and callback execution
// upon completion. It provides go.h compatible HTML.
package loader

import (
	"encoding/json"
	"github.com/daaku/go.h"
	"log"
)

// The global loader name.
const LoaderName = "$L"

// Each loadable resource must be defined in terms of URLs and some
// optional JavaScript.
type Resource interface {
	URLs() []string
	Script() string
}

// Defines the HTML tag including the resources to load.
type HTML struct {
	Resource []Resource
}

func (l *HTML) HTML() (h.HTML, error) {
	if len(l.Resource) == 0 {
		return nil, nil
	}

	script := ""
	for _, t := range l.Resource {
		script += getScript(t)
	}
	return &h.Script{
		Inner: h.Unsafe(LoaderName + "=" +
			`function(c,d){for(var b=c.length,e=b,f=function(){if(!(this.readyState` +
			`&&this.readyState!=="complete"&&this.readyState!=="loaded")){` +
			`this.onload= this.onreadystatechange=null;--e||d()}},g=` +
			`document.getElementsByTagName("head")[0],i=function(h){` +
			`var a=document.createElement("script");a.async= true;a.src=h;` +
			`a.onload=a.onreadystatechange=f;g.appendChild(a)};b;)i(c[--b])};` +
			script),
	}, nil
}

func getScript(l Resource) string {
	urls := l.URLs()
	encoded, err := json.Marshal(urls)
	if err != nil {
		log.Fatalf("Failed to encode URLs string array %v with error %s", urls, err)
	}
	return LoaderName + "(" + string(encoded) + ", function() {" + l.Script() + "});"
}
