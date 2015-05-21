// Command rell implements the main webserver application for Rell.
package main

import (
	"flag"
	"fmt"
	"go/build"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/daaku/rell/adminweb"
	"github.com/daaku/rell/examples"
	"github.com/daaku/rell/examples/viewexamples"
	"github.com/daaku/rell/internal/github.com/GeertJohan/go.rice"
	"github.com/daaku/rell/internal/github.com/daaku/go.browserid"
	"github.com/daaku/rell/internal/github.com/daaku/go.static"
	"github.com/daaku/rell/internal/github.com/daaku/go.trustforward"
	"github.com/daaku/rell/internal/github.com/daaku/go.xsrf"
	"github.com/daaku/rell/internal/github.com/facebookgo/devrestarter"
	"github.com/daaku/rell/internal/github.com/facebookgo/fbapi"
	"github.com/daaku/rell/internal/github.com/facebookgo/fbapp"
	"github.com/daaku/rell/internal/github.com/facebookgo/flagenv"
	"github.com/daaku/rell/internal/github.com/facebookgo/httpcontrol"
	"github.com/daaku/rell/internal/github.com/facebookgo/httpdown"
	"github.com/daaku/rell/internal/github.com/facebookgo/parse"
	"github.com/daaku/rell/internal/github.com/golang/groupcache/lru"
	"github.com/daaku/rell/oauth"
	"github.com/daaku/rell/og"
	"github.com/daaku/rell/og/viewog"
	"github.com/daaku/rell/rellenv"
	"github.com/daaku/rell/rellenv/appns"
	"github.com/daaku/rell/rellenv/empcheck"
	"github.com/daaku/rell/rellenv/viewcontext"
	"github.com/daaku/rell/web"
)

func defaultAddr() string {
	if port := os.Getenv("PORT"); port != "" {
		return ":" + port
	}
	return ":43600"
}

func pkgDir(path string) string {
	pkg, err := build.Import(path, "", build.FindOnly)
	if err != nil {
		return ""
	}
	return pkg.Dir
}

func main() {
	const signedRequestMaxAge = time.Hour * 24
	runtime.GOMAXPROCS(runtime.NumCPU())

	flagSet := flag.NewFlagSet(filepath.Base(os.Args[0]), flag.ExitOnError)
	dev := flagSet.Bool("dev", runtime.GOOS != "linux", "development mode")
	addr := flagSet.String("addr", defaultAddr(), "server address to bind to")
	adminPath := flagSet.String("admin-path", "", "secret admin path")
	facebookAppID := flagSet.Uint64("fb-app-id", 342526215814610, "facebook application id")
	facebookAppSecret := flagSet.String("fb-app-secret", "", "facebook application secret")
	facebookAppNS := flagSet.String("fb-app-ns", "", "facebook application namespace")
	empCheckerAppID := flagSet.Uint64("empcheck-app-id", 0, "empcheck application id")
	empCheckerAppSecret := flagSet.String("empcheck-app-secret", "", "empcheck application secret")
	parseAppID := flagSet.String("parse-app-id", "", "parse application id")
	parseMasterKey := flagSet.String("parse-master-key", "", "parse master key")
	publicDir := flagSet.String(
		"public-dir", pkgDir("github.com/daaku/rell/public"), "public files directory")

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
		X:          true,
		CloudFlare: true,
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
	publicFS := http.Dir(*publicDir)
	static := &static.Handler{
		Path: "/static/",
		Box:  static.FileSystemBox(publicFS),
	}
	httpTransport := &httpcontrol.Transport{
		MaxIdleConnsPerHost:   http.DefaultMaxIdleConnsPerHost,
		DialTimeout:           2 * time.Second,
		ResponseHeaderTimeout: 3 * time.Second,
		RequestTimeout:        30 * time.Second,
	}
	parseClient := &parse.Client{
		Transport: httpTransport,
		Credentials: parse.MasterKey{
			ApplicationID: *parseAppID,
			MasterKey:     *parseMasterKey,
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
	webHandler := &web.Handler{
		Static: static,
		App:    fbApp,
		Logger: logger,
		EnvParser: &rellenv.Parser{
			App:                 fbApp,
			EmpChecker:          empChecker,
			AppNSFetcher:        appNSFetcher,
			SignedRequestMaxAge: signedRequestMaxAge,
			Forwarded:           forwarded,
		},
		PublicFS:       publicFS,
		ContextHandler: &viewcontext.Handler{},
		ExamplesHandler: &viewexamples.Handler{
			ExampleStore: exampleStore,
			Xsrf:         xsrf,
			Static:       static,
		},
		OgHandler: &viewog.Handler{
			Static:       static,
			ObjectParser: &og.Parser{Static: static},
		},
		OauthHandler: &oauth.Handler{
			BrowserID:     bid,
			App:           fbApp,
			HttpTransport: httpTransport,
			Static:        static,
		},
		AdminHandler: &adminweb.Handler{
			Forwarded: forwarded,
			Path:      *adminPath,
			SkipHTTPS: *dev,
		},
		SignedRequestMaxAge: signedRequestMaxAge,
	}
	httpServer := &http.Server{
		Addr:    *addr,
		Handler: webHandler,
	}
	hdConfig := &httpdown.HTTP{
		StopTimeout: 9 * time.Second, // heroku provides 10 seconds to terminate
	}

	if err := httpdown.ListenAndServe(httpServer, hdConfig); err != nil {
		logger.Fatal(err)
	}
}
