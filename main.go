// Command rell implements the main webserver application for Rell.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/GeertJohan/go.rice"
	"github.com/daaku/go.browserid"
	"github.com/daaku/go.static"
	"github.com/daaku/go.xsrf"
	"github.com/facebookgo/fbapi"
	"github.com/facebookgo/fbapp"
	"github.com/facebookgo/flagconfig"
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
	// TODO: switch to a local flagset and patch flagenv & flagconfig to allow
	// taking in a FlagSet.

	var f flags
	flag.StringVar(&f.Addr, "addr", ":43600", "server address to bind to")
	flag.StringVar(&f.AdminAddr, "admin-addr", ":43601", "admin http server address")
	flag.Uint64Var(&f.FacebookAppID, "fb-app-id", 342526215814610, "facebook application id")
	flag.StringVar(&f.FacebookAppSecret, "fb-app-secret", "", "facebook application secret")
	flag.StringVar(&f.FacebookAppNS, "fb-app-ns", "", "facebook application namespace")
	flag.Uint64Var(&f.EmpCheckerAppID, "empcheck-app-id", 0, "empcheck application id")
	flag.StringVar(&f.EmpCheckerAppSecret, "empcheck-app-secret", "", "empcheck application secret")
	flag.StringVar(&f.ParseAppID, "parse-app-id", "", "parse application id")
	flag.StringVar(&f.ParseRestAPIKey, "parse-rest-api-key", "", "parse rest api key")

	flag.Usage = flagconfig.Usage
	flag.Parse()
	flagenv.Parse()
	flagconfig.Parse()

	return f
}

func main() {
	flags := globalFlags()
	logger := log.New(os.Stderr, "", log.LstdFlags)
	mainapp := fbapp.New(
		flags.FacebookAppID,
		flags.FacebookAppSecret,
		flags.FacebookAppNS,
	)
	bid := &browserid.Cookie{
		Name:   "z",
		MaxAge: time.Hour * 24 * 365 * 10, // 10 years
		Length: 16,
		Logger: logger,
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
		Transport:     httpTransport,
		ApplicationID: flags.ParseAppID,
		Credentials:   parse.RestAPIKey(flags.ParseRestAPIKey),
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
		App:          mainapp,
		EmpChecker:   empChecker,
		AppNSFetcher: appNSFetcher,
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
