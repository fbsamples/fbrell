package browserify_test

import (
	"github.com/daaku/go.browserify"
	"go/build"
	"log"
	"strings"
	"testing"
)

const example = "github.com/daaku/go.browserify/example"

var s = &browserify.Script{
	Dir:   exampleDir(),
	Entry: "lib/example.js",
}

func exampleDir() string {
	pkg, err := build.Import(example, "", build.FindOnly)
	if err != nil {
		log.Fatalf("Failed to find example npm module: %s", err)
	}
	return pkg.Dir
}

func TestContents(t *testing.T) {
	b, err := s.Content()
	if err != nil {
		t.Fatalf("Error getting content: %s", err)
	}
	content := string(b)
	if !strings.Contains(content, "dotaccess") {
		t.Fatalf("Was expecting dotaccess but did not find it:\n%s", content)
	}
	if !strings.Contains(content, "42 + ans") {
		t.Fatalf("Was expecting example content but did not find it:\n%s", content)
	}
}

func TestURL(t *testing.T) {
	const expected = "/browserify/d5b99e1aba/lib/example.js"
	u, err := s.URL()
	if err != nil {
		t.Fatalf("Error getting URL: %s", err)
	}
	if u != expected {
		t.Fatalf("Did not find expected URL %s instead found %s", expected, u)
	}
}
