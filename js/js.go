// Package js provides the Rell JavaScript resources.
package js

import (
	"encoding/json"
	"fmt"
	"github.com/nshah/go.browserify"
	"github.com/nshah/rell/context"
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
}

func (i *Init) URLs() []string {
	url, err := defaultScript.URL()
	if err != nil {
		log.Fatalf("Failed to get browserify script URL: %s", err)
	}
	return []string{url, i.Context.SdkURL()}
}

func (i *Init) Script() string {
	encoded, err := json.Marshal(i.Context)
	if err != nil {
		log.Fatalf("Failed to json.Marshal context: %s", err)
	}
	return fmt.Sprintf("require('./rell').init(%s)", string(encoded))
}
