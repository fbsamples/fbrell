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

	"github.com/GeertJohan/go.rice"
	"github.com/daaku/go.browserid"
	"github.com/daaku/go.static"
	"github.com/daaku/go.trustforward"
	"github.com/daaku/go.xsrf"
	"github.com/facebookgo/devrestarter"
	"github.com/facebookgo/fbapi"
	"github.com/facebookgo/fbapp"
	"github.com/facebookgo/flagenv"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/facebookgo/httpcontrol"
	"github.com/facebookgo/parse"
	"github.com/golang/groupcache/lru"

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

type flags struct {
	Dev                 bool
	Addr                string
	AdminAddr           string
	FacebookAppID       uint64
	FacebookAppSecret   string
	FacebookAppNS       string
	EmpCheckerAppID     uint64
	EmpCheckerAppSecret string
	ParseAppID          string
	ParseRestAPIKey     string
}

func globalFlags() flags {
	set := flag.NewFlagSet(filepath.Base(os.Args[0]), flag.ExitOnError)
	f := flags{}

	set.BoolVar(&f.Dev, "dev", runtime.GOOS != "linux", "development mode")
	set.StringVar(&f.Addr, "addr", ":43600", "server address to bind to")
	set.StringVar(&f.AdminAddr, "admin-addr", ":43601", "admin http server address")
	set.Uint64Var(&f.FacebookAppID, "fb-app-id", 342526215814610, "facebook application id")
	set.StringVar(&f.FacebookAppSecret, "fb-app-secret", "", "facebook application secret")
	set.StringVar(&f.FacebookAppNS, "fb-app-ns", "", "facebook application namespace")
	set.Uint64Var(&f.EmpCheckerAppID, "empcheck-app-id", 0, "empcheck application id")
	set.StringVar(&f.EmpCheckerAppSecret, "empcheck-app-secret", "", "empcheck application secret")
	set.StringVar(&f.ParseAppID, "parse-app-id", "", "parse application id")
	set.StringVar(&f.ParseRestAPIKey, "parse-rest-api-key", "", "parse rest api key")

	set.Parse(os.Args[1:])
	if err := flagenv.ParseSet("RELL_", set); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	return f
}

func main() {
	const signedRequestMaxAge = time.Hour * 24
	flags := globalFlags()

	if flags.Dev {
		devrestarter.Init()
	}

	logger := log.New(os.Stderr, "", log.LstdFlags)
	mainapp := fbapp.New(
		flags.FacebookAppID,
		flags.FacebookAppSecret,
		flags.FacebookAppNS,
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
			ApplicationID: flags.ParseAppID,
			RestAPIKey:    flags.ParseRestAPIKey,
		},
	}
	fbApiClient := &fbapi.Client{
		Redact:    true,
		Transport: httpTransport,
	}
	lruCache := lru.New(10000)
	empChecker := &empcheck.Checker{
		FbApiClient: fbApiClient,
		App:         fbapp.New(flags.EmpCheckerAppID, flags.EmpCheckerAppSecret, ""),
		Logger:      logger,
		Cache:       lruCache,
	}
	appNSFetcher := &appns.Fetcher{
		Apps:        []fbapp.App{mainapp},
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
		App:                 mainapp,
		EmpChecker:          empChecker,
		AppNSFetcher:        appNSFetcher,
		SignedRequestMaxAge: signedRequestMaxAge,
		Forwarded:           forwarded,
	}

	app := &web.App{
		Static: static,
		App:    mainapp,
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
			App:           mainapp,
			ContextParser: contextParser,
			HttpTransport: httpTransport,
			Static:        static,
		},
		SignedRequestMaxAge: signedRequestMaxAge,
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	// for systemd started servers we can skip the date/time since journald
	// already shows it
	if os.Getppid() == 1 {
		logger.SetFlags(0)
	}

	err := gracehttp.Serve(
		&http.Server{Addr: flags.Addr, Handler: http.HandlerFunc(app.MainHandler)},
		&http.Server{Addr: flags.AdminAddr, Handler: http.HandlerFunc(app.AdminHandler)},
	)
	if err != nil {
		logger.Fatal(err)
	}
}
