// Package js provides the Rell JavaScript resources.
package js

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/nshah/go.browserify"
	"github.com/nshah/go.flag.pkgpath"
	"github.com/nshah/rell/context"
	"github.com/nshah/rell/examples"
	"log"
)

// The default script. The working directory is set in init().
var defaultScript = &browserify.Script{Entry: "rell.js"}

func init() {
	pkgpath.DirVar(
		&defaultScript.Dir,
		"rell.browserify.dir",
		"github.com/nshah/rell/js",
		"The browserify working directory.")
	flag.StringVar(
		&defaultScript.Override,
		"rell.browserify.override",
		"",
		"Pre-generated browserify output file.")
}

// Represents configuration for initializing the rell
// module. Essentiall does a "require("./rell").init(x...)" call.
type Init struct {
	Context *context.Context
	Example *examples.Example
}

func (i *Init) URLs() []string {
	url, err := defaultScript.URL()
	if err != nil {
		log.Fatalf("Failed to get browserify script URL: %s", err)
	}
	return []string{url, i.Context.SdkURL()}
}

func (i *Init) Script() string {
	encodedContext, err := json.Marshal(i.Context)
	if err != nil {
		log.Fatalf("Failed to json.Marshal context: %s", err)
	}
	encodedExample, err := json.Marshal(i.Example)
	if err != nil {
		log.Fatalf("Failed to json.Marshal example: %s", err)
	}
	return fmt.Sprintf(
		"require('./rell').init(%s, %s)", encodedContext, encodedExample)
}
