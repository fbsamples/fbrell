// Package js provides the Rell JavaScript resources.
package js

import (
	"encoding/json"
	"fmt"
	"github.com/nshah/go.browserify"
	"github.com/nshah/rell/context"
	"github.com/nshah/rell/examples"
	"go/build"
	"log"
)

// The default script.
var defaultScript = &browserify.Script{
	Dir:   getBrowserifyDir(),
	Entry: "rell.js",
}

func init() {
	defaultScript.URL() // called for side-effect of priming the internal cache
}

// Find the package relative directory to use as the browserify directory.
func getBrowserifyDir() string {
	pkg, err := build.Import(
		"github.com/nshah/rell/public", "", build.FindOnly)
	if err != nil {
		log.Fatalf("Failed to find browserify directory: %s", err)
	}
	return pkg.Dir
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
		"require('./rell').init(%s, %s)",
		string(encodedContext),
		string(encodedExample))
}
