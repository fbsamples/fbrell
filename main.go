/**
 * Copyright (c) 2014-present, Facebook, Inc. All rights reserved.
 *
 * You are hereby granted a non-exclusive, worldwide, royalty-free license to use,
 * copy, modify, and distribute this software in source code or binary form for use
 * in connection with the web services and APIs provided by Facebook.
 *
 * As with any software that integrates with the Facebook platform, your use of
 * this software is subject to the Facebook Developer Principles and Policies
 * [http://developers.facebook.com/policy/]. This copyright notice shall be
 * included in all copies or substantial portions of the software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

// Command rell implements the main webserver application for Rell.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	browserid "github.com/daaku/go.browserid"
	static "github.com/daaku/go.static"
	trustforward "github.com/daaku/go.trustforward"
	"github.com/facebookgo/devrestarter"
	"github.com/facebookgo/fbapi"
	"github.com/facebookgo/fbapp"
	"github.com/facebookgo/flagenv"
	"github.com/facebookgo/httpcontrol"
	"github.com/facebookgo/httpdown"
	"github.com/fbsamples/fbrell/adminweb"
	"github.com/fbsamples/fbrell/examples"
	"github.com/fbsamples/fbrell/examples/viewexamples"
	"github.com/fbsamples/fbrell/oauth"
	"github.com/fbsamples/fbrell/og"
	"github.com/fbsamples/fbrell/og/viewog"
	"github.com/fbsamples/fbrell/rellenv"
	"github.com/fbsamples/fbrell/rellenv/appns"
	"github.com/fbsamples/fbrell/rellenv/empcheck"
	"github.com/fbsamples/fbrell/rellenv/viewcontext"
	"github.com/fbsamples/fbrell/web"
	"github.com/golang/groupcache/lru"
)

func defaultAddr() string {
	if port := os.Getenv("PORT"); port != "" {
		return ":" + port
	}
	return ":43600"
}

func main() {
	const signedRequestMaxAge = time.Hour * 24

	dev := flag.Bool("dev", runtime.GOOS != "linux", "development mode")
	addr := flag.String("addr", defaultAddr(), "server address to bind to")
	adminPath := flag.String("admin-path", "", "secret admin path")
	facebookAppID := flag.Uint64("fb-app-id", 342526215814610, "facebook application id")
	facebookAppSecret := flag.String("fb-app-secret", "", "facebook application secret")
	facebookAppNS := flag.String("fb-app-ns", "", "facebook application namespace")
	empCheckerAppID := flag.Uint64("empcheck-app-id", 0, "empcheck application id")
	empCheckerAppSecret := flag.String("empcheck-app-secret", "", "empcheck application secret")
	publicDir := flag.String(
		"public-dir", "./public", "public files directory")
	examplesDir := flag.String(
		"examples-dir", "./examples/db", "example files directory")

	flag.Parse()
	if err := flagenv.ParseSet("RELL_", flag.CommandLine); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if *dev {
		devrestarter.Init()
	}

	logger := log.New(os.Stderr, "", log.LstdFlags)
	// for systemd started servers we can skip the date/time since journald
	// already shows it
	if os.Getenv("JOURNAL_STREAM") != "" {
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
	fbApiClient := &fbapi.Client{
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
		DB: examples.MustMakeDB(*examplesDir),
	}
	adminHandler := &adminweb.Handler{
		Forwarded: forwarded,
		Path:      *adminPath,
		SkipHTTPS: *dev,
	}
	adminHandler.Init()
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
		AdminHandler: adminHandler,
		SignedRequestMaxAge: signedRequestMaxAge,
	}
	if err := webHandler.Init(); err != nil {
		logger.Fatal(err)
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
