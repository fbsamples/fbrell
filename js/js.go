// Package js provides the Rell JavaScript resources.
package js

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/daaku/rell/context"
	"github.com/daaku/rell/examples"
)

// Represents configuration for initializing the rell
// module. Essentiall does a "require("./rell").init(x...)" call.
type Init struct {
	Context *context.Context
	Example *examples.Example
}

func (i *Init) URLs() []string {
	return []string{i.Context.SdkURL()}
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
	return fmt.Sprintf("Rell.init(%s, %s)", encodedContext, encodedExample)
}
