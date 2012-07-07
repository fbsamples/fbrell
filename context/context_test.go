package context_test

import (
	"github.com/daaku/go.fbapp"
	"github.com/daaku/go.subset"
	"github.com/daaku/rell/context"
	"net/http"
	"net/url"
	"testing"
)

func init() {
	fbapp.Default = fbapp.New(184484190795, "", "fbrelll")
}

func fromValues(t *testing.T, values url.Values) *context.Context {
	req, err := http.NewRequest(
		"GET",
		"http://www.fbrell.com/?"+values.Encode(),
		nil)
	if err != nil {
		t.Fatalf("Failed to create request: %s", err)
	}
	ctx, err := context.FromRequest(req)
	if err != nil {
		t.Fatalf("Failed to create context: %s", err)
	}
	return ctx
}

func TestDefaultContext(t *testing.T) {
	t.Parallel()
	ctx := fromValues(t, url.Values{})
	subset.Assert(t, context.Default(), ctx)
}

func TestCustomAppID(t *testing.T) {
	t.Parallel()
	values := url.Values{}
	values.Add("appid", "123")
	context := fromValues(t, values)
	if context.AppID != 123 {
		t.Fatalf("Did not find expected app id 123 instead found %d", context.AppID)
	}
}

func TestCustomStatus(t *testing.T) {
	t.Parallel()
	values := url.Values{}
	values.Add("status", "0")
	context := fromValues(t, values)
	if context.Status {
		t.Fatal("Was expecting status to be false.")
	}
}

func TestComplex(t *testing.T) {
	t.Parallel()
	values := url.Values{}
	values.Add("status", "1")
	values.Add("server", "beta")
	values.Add("locale", "en_PI")
	values.Add("version", "old")
	values.Add("channel", "false")
	expected := &context.Context{
		Status:     true,
		Env:        "beta",
		Locale:     "en_PI",
		Version:    context.Old,
		UseChannel: false,
	}
	context := fromValues(t, values)
	subset.Assert(t, expected, context)
}

func TestPageTabURLBeta(t *testing.T) {
	t.Parallel()
	expected := "http://www.beta.facebook.com/pages/" +
		"Rell-Page-for-Tabs/141929622497380?sk=app_184484190795" +
		"&app_data=Lz9zZXJ2ZXI9YmV0YQ%3D%3D"
	values := url.Values{}
	values.Add("server", "beta")
	context := fromValues(t, values)
	actual := context.PageTabURL("/")
	if actual != expected {
		t.Fatalf("Did not find expected URL %s instead found %s", expected, actual)
	}
}

func TestPageTabURL(t *testing.T) {
	t.Parallel()
	expected := "http://www.facebook.com/pages/Rell-Page-for-Tabs" +
		"/141929622497380?sk=app_184484190795&app_data=Lw%3D%3D"
	context := fromValues(t, url.Values{})
	if context.PageTabURL("/") != expected {
		t.Fatalf("Did not find expected URL %s instead found %s",
			expected, context.PageTabURL("/"))
	}
}

func TestCanvasURLBeta(t *testing.T) {
	t.Parallel()
	expected := "http://apps.beta.facebook.com/fbrelll/?server=beta"
	values := url.Values{}
	values.Add("server", "beta")
	context := fromValues(t, values)
	if context.CanvasURL("/") != expected {
		t.Fatalf("Did not find expected URL %s instead found %s",
			expected, context.CanvasURL("/"))
	}
}

func TestCanvasURL(t *testing.T) {
	t.Parallel()
	expected := "http://apps.facebook.com/fbrelll/"
	context := fromValues(t, url.Values{})
	if context.CanvasURL("/") != expected {
		t.Fatalf("Did not find expected URL %s instead found %s",
			expected, context.CanvasURL("/"))
	}
}
