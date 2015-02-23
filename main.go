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
	"github.com/daaku/go.redis"
	"github.com/daaku/go.redis/bytestore"
	"github.com/daaku/go.static"
	"github.com/daaku/go.xsrf"
	"github.com/facebookgo/fbapi"
	"github.com/facebookgo/fbapp"
	"github.com/facebookgo/flagconfig"
	"github.com/facebookgo/flagenv"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/facebookgo/httpcontrol"
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

func main() {
	logger := log.New(os.Stderr, "", log.LstdFlags)
	mainapp := fbapp.Flag("fbapp")
	bid := &browserid.Cookie{
		Name:   "z",
		MaxAge: time.Hour * 24 * 365 * 10, // 10 years
		Length: 16,
		Logger: logger,
	}
	redis := redis.ClientFlag("rell.redis")
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
	byteStore := bytestore.New(redis)
	httpTransport := &httpcontrol.Transport{
		MaxIdleConnsPerHost:   http.DefaultMaxIdleConnsPerHost,
		DialTimeout:           2 * time.Second,
		ResponseHeaderTimeout: 3 * time.Second,
		RequestTimeout:        30 * time.Second,
	}
	fbApiClient := &fbapi.Client{
		Redact:    true,
		Transport: httpTransport,
	}
	lruCache := lru.New(10000)
	empChecker := &empcheck.Checker{
		FbApiClient: fbApiClient,
		App:         fbapp.Flag("empcheck"),
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
		ByteStore: byteStore,
		DB:        examples.MustMakeDB(rice.MustFindBox("examples/db")),
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

	mainAddress := flag.String(
		"addr",
		":43600",
		"Server address to bind to.",
	)
	adminAddress := flag.String(
		"admin-addr",
		":43601",
		"Admin http server address.",
	)

	flag.Usage = flagconfig.Usage
	flag.Parse()
	flagenv.Parse()
	flagconfig.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	// for systemd started servers we can skip the date/time since journald
	// already shows it
	if os.Getppid() == 1 {
		logger.SetFlags(0)
	}

	err := gracehttp.Serve(
		&http.Server{Addr: *mainAddress, Handler: http.HandlerFunc(app.MainHandler)},
		&http.Server{Addr: *adminAddress, Handler: http.HandlerFunc(app.AdminHandler)},
	)
	if err != nil {
		logger.Fatal(err)
	}
}
