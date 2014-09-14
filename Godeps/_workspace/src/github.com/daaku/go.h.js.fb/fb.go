// Package fb provides go.h compatible async loading for the Facebook JS SDK.
package fb

import (
	"encoding/json"
	"fmt"
	"log"
)

// Represents an async load and FB.init call for the Facebook JS SDK.
type Init struct {
	URL   string `json:"-"`
	AppID uint64 `json:"appId"`
}

const defaultURL = "//connect.facebook.net/en_US/all.js"

func (i *Init) URLs() []string {
	url := i.URL
	if url == "" {
		url = defaultURL
	}
	return []string{url}
}

func (i *Init) Script() string {
	encoded, err := json.Marshal(i)
	if err != nil {
		log.Fatalf("Failed to json.Marshal sdk.Init %+v with error %s", i, err)
	}
	return fmt.Sprintf("FB.init(%s)", string(encoded))
}
