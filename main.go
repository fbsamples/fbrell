// Command rell implements the main webserver application for Rell.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/daaku/rell/Godeps/_workspace/src/github.com/GeertJohan/go.rice"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.browserid"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.static"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.trustforward"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.xsrf"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/facebookgo/devrestarter"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/facebookgo/fbapi"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/facebookgo/fbapp"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/facebookgo/flagenv"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/facebookgo/grace/gracehttp"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/facebookgo/httpcontrol"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/facebookgo/parse"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/golang/groupcache/lru"
	"github.com/daaku/rell/context"
	"github.com/daaku/rell/context/appns"
	"github.com/daaku/rell/context/empcheck"
	"github.com/daaku/rell/context/viewcontext"
	"github.com/daaku/rell/examples"
	"github.com/daaku/rell/examples/viewexamples"
	"github.com/daaku/rell/oauth"
	"github.com/daaku/rell/og"
	"github.com/daaku/rell/og/viewog"
	"github.com/daaku/rell/web"
)

func main() {
	const signedRequestMaxAge = time.Hour * 24
	runtime.GOMAXPROCS(runtime.NumCPU())

	flagSet := flag.NewFlagSet(filepath.Base(os.Args[0]), flag.ExitOnError)
	dev := flagSet.Bool("dev", runtime.GOOS != "linux", "development mode")
	addr := flagSet.String("addr", ":43600", "server address to bind to")
	adminAddr := flagSet.String("admin-addr", ":43601", "admin http server address")
	facebookAppID := flagSet.Uint64("fb-app-id", 342526215814610, "facebook application id")
	facebookAppSecret := flagSet.String("fb-app-secret", "", "facebook application secret")
	facebookAppNS := flagSet.String("fb-app-ns", "", "facebook application namespace")
	empCheckerAppID := flagSet.Uint64("empcheck-app-id", 0, "empcheck application id")
	empCheckerAppSecret := flagSet.String("empcheck-app-secret", "", "empcheck application secret")
	parseAppID := flagSet.String("parse-app-id", "", "parse application id")
	parseRestAPIKey := flagSet.String("parse-rest-api-key", "", "parse rest api key")

	flagSet.Parse(os.Args[1:])
	if err := flagenv.ParseSet("RELL_", flagSet); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if *dev {
		devrestarter.Init()
	}

	logger := log.New(os.Stderr, "", log.LstdFlags)
	// for systemd started servers we can skip the date/time since journald
	// already shows it
	if os.Getppid() == 1 {
		logger.SetFlags(0)
	}

	fbApp := fbapp.New(
		*facebookAppID,
		*facebookAppSecret,
		*facebookAppNS,
	)
	forwarded := &trustforward.Forwarded{
		X: true,
	}
	bid := &browserid.Cookie{
		Name:      "z",
		MaxAge:    time.Hour * 24 * 365 * 10, // 10 years
		Length:    16,
		Logger:    logger,
		Forwarded: forwarded,
	}
	xsrf := &xsrf.Provider{
		BrowserID: bid,
		MaxAge:    time.Hour * 24,
		SumLen:    10,
	}
	static := &static.Handler{
		HttpPath:    "/static/",
		MaxAge:      time.Hour * 24 * 365,
		MemoryCache: true,
		Box:         rice.MustFindBox("public"),
	}
	httpTransport := &httpcontrol.Transport{
		MaxIdleConnsPerHost:   http.DefaultMaxIdleConnsPerHost,
		DialTimeout:           2 * time.Second,
		ResponseHeaderTimeout: 3 * time.Second,
		RequestTimeout:        30 * time.Second,
	}
	parseClient := &parse.Client{
		Transport: httpTransport,
		Credentials: parse.RestAPIKey{
			ApplicationID: *parseAppID,
			RestAPIKey:    *parseRestAPIKey,
		},
	}
	fbApiClient := &fbapi.Client{
		Redact:    true,
		Transport: httpTransport,
	}
	lruCache := lru.New(10000)
	empChecker := &empcheck.Checker{
		FbApiClient: fbApiClient,
		App:         fbapp.New(*empCheckerAppID, *empCheckerAppSecret, ""),
		Logger:      logger,
		Cache:       lruCache,
	}
	appNSFetcher := &appns.Fetcher{
		Apps:        []fbapp.App{fbApp},
		FbApiClient: fbApiClient,
		Logger:      logger,
		Cache:       lruCache,
	}
	exampleStore := &examples.Store{
		Parse: parseClient,
		DB:    examples.MustMakeDB(rice.MustFindBox("examples/db")),
		Cache: lruCache,
	}
	contextParser := &context.Parser{
		App:                 fbApp,
		EmpChecker:          empChecker,
		AppNSFetcher:        appNSFetcher,
		SignedRequestMaxAge: signedRequestMaxAge,
		Forwarded:           forwarded,
	}

	app := &web.App{
		Static: static,
		App:    fbApp,
		ContextHandler: &viewcontext.Handler{
			ContextParser: contextParser,
			Static:        static,
		},
		ExamplesHandler: &viewexamples.Handler{
			ContextParser: contextParser,
			ExampleStore:  exampleStore,
			Xsrf:          xsrf,
			Static:        static,
		},
		OgHandler: &viewog.Handler{
			ContextParser: contextParser,
			Static:        static,
			ObjectParser:  &og.Parser{Static: static},
		},
		OauthHandler: &oauth.Handler{
			BrowserID:     bid,
			App:           fbApp,
			ContextParser: contextParser,
			HttpTransport: httpTransport,
			Static:        static,
		},
		SignedRequestMaxAge: signedRequestMaxAge,
	}

	err := gracehttp.Serve(
		&http.Server{Addr: *addr, Handler: http.HandlerFunc(app.MainHandler)},
		&http.Server{Addr: *adminAddr, Handler: http.HandlerFunc(app.AdminHandler)},
	)
	if err != nil {
		logger.Fatal(err)
	}
}
