package context_test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.trustforward"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/facebookgo/ensure"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/facebookgo/fbapp"
	"github.com/daaku/rell/context"
)

const (
	defaultFacebookAppID = 42
	defaultAppNS         = "fbrelll"
)

type funcEmpChecker func(uint64) bool

func (f funcEmpChecker) Check(uid uint64) bool {
	return f(uid)
}

type funcAppNSFetcher func(uint64) string

func (f funcAppNSFetcher) Get(id uint64) string {
	return f(id)
}

func defaultParser() *context.Parser {
	return &context.Parser{
		EmpChecker:   funcEmpChecker(func(uint64) bool { return true }),
		AppNSFetcher: funcAppNSFetcher(func(uint64) string { return defaultAppNS }),
		App:          fbapp.New(defaultFacebookAppID, "", ""),
		Forwarded:    &trustforward.Forwarded{},
	}
}

func fromValues(t *testing.T, values url.Values) *context.Context {
	req, err := http.NewRequest(
		"GET",
		"http://www.fbrell.com/?"+values.Encode(),
		nil)
	if err != nil {
		t.Fatalf("Failed to create request: %s", err)
	}
	ctx, err := defaultParser().FromRequest(req)
	if err != nil {
		t.Fatalf("Failed to create context: %s", err)
	}
	return ctx
}

func TestDefaultContext(t *testing.T) {
	t.Parallel()
	ctx := fromValues(t, url.Values{})
	ensure.Subset(t, ctx, defaultParser().Default())
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
	expected := &context.Context{
		Status: true,
		Env:    "beta",
		Locale: "en_PI",
	}
	context := fromValues(t, values)
	ensure.Subset(t, context, expected)
}

func TestPageTabURLBeta(t *testing.T) {
	t.Parallel()
	context := fromValues(t, url.Values{"server": []string{"beta"}})
	pageTabURL := context.PageTabURL("/")
	ensure.StringContains(t, pageTabURL,
		"http://www.beta.facebook.com/pages/Rell-Page-for-Tabs/141929622497380")
}

func TestPageTabURL(t *testing.T) {
	t.Parallel()
	context := fromValues(t, url.Values{})
	pageTabURL := context.PageTabURL("/")
	ensure.StringContains(t, pageTabURL,
		"http://www.facebook.com/pages/Rell-Page-for-Tabs/141929622497380")
	ensure.StringContains(t, pageTabURL, fmt.Sprintf("app_%d", defaultFacebookAppID))
	ensure.StringContains(t, pageTabURL, "app_data=Lw%3D%3D")
}

func TestCanvasURLBeta(t *testing.T) {
	t.Parallel()
	context := fromValues(t, url.Values{"server": []string{"beta"}})
	canvasURL := context.CanvasURL("/")
	ensure.StringContains(t, canvasURL,
		fmt.Sprintf("https://apps.beta.facebook.com/%s/?server=beta", defaultAppNS))
}

func TestCanvasURL(t *testing.T) {
	t.Parallel()
	context := fromValues(t, url.Values{})
	canvasURL := context.CanvasURL("/")
	ensure.StringContains(t, canvasURL,
		fmt.Sprintf("https://apps.facebook.com/%s/", defaultAppNS))
}
