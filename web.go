// Command rell implements the main webserver application for Rell.
package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"path/filepath"
	"runtime"

	"github.com/daaku/go.browserify"
	"github.com/daaku/go.fbapp"
	"github.com/daaku/go.flagconfig"
	"github.com/daaku/go.grace/gracehttp"
	"github.com/daaku/go.httpdev"
	"github.com/daaku/go.httpgzip"
	"github.com/daaku/go.httpstats"
	"github.com/daaku/go.signedrequest/appdata"
	"github.com/daaku/go.viewvar"

	"github.com/daaku/rell/context/viewcontext"
	"github.com/daaku/rell/examples/viewexamples"
	"github.com/daaku/rell/oauth"
	"github.com/daaku/rell/og/viewog"
	"github.com/daaku/rell/service"
)

var (
	mainAddress = flag.String(
		"rell.address",
		":43600",
		"Server address to bind to.",
	)
	adminAddress = flag.String(
		"rell.admin.address",
		":43601",
		"Admin http server address.",
	)
	goMaxProcs = flag.Int(
		"rell.gomaxprocs",
		runtime.NumCPU(),
		"Maximum processes to use.",
	)
)

func main() {
	flag.Usage = flagconfig.Usage
	flag.Parse()
	flagconfig.Parse()
	runtime.GOMAXPROCS(*goMaxProcs)

	if err := service.Stats.Start(); err != nil {
		log.Fatal(err)
	}

	// for systemd started servers we can skip the date/time since journald
	// already shows it
	if os.Getppid() == 1 {
		log.SetFlags(0)
		service.Logger.SetFlags(0)
	}

	err := gracehttp.Serve(
		&http.Server{Addr: *mainAddress, Handler: mainHandler()},
		&http.Server{Addr: *adminAddress, Handler: adminHandler()},
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := service.Stats.Stop(); err != nil {
		log.Fatal(err)
	}
}

// binds a path to a single file
func staticFile(mux *http.ServeMux, name string) {
	abs := filepath.Join(service.Static.DiskPath, name)
	mux.HandleFunc(name, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, abs)
	})
}

func adminHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/vars/", viewvar.Json)
	return mux
}

func mainHandler() (handler http.Handler) {
	const public = "/public/"

	mux := http.NewServeMux()
	mux.Handle(service.Static.HttpPath, service.Static)

	staticFile(mux, "/favicon.ico")
	staticFile(mux, "/f8.jpg")
	staticFile(mux, "/robots.txt")
	mux.Handle(public,
		http.StripPrefix(
			public, http.FileServer(http.Dir(service.Static.DiskPath))))

	mux.HandleFunc(browserify.Path, browserify.Handle)
	mux.HandleFunc("/not_a_real_webpage", http.NotFound)
	mux.HandleFunc("/info/", viewcontext.Info)
	mux.HandleFunc("/examples/", viewexamples.List)
	mux.HandleFunc("/saved/", viewexamples.Saved)
	mux.HandleFunc("/raw/", viewexamples.Raw)
	mux.HandleFunc("/simple/", viewexamples.Simple)
	mux.HandleFunc("/channel/", viewexamples.SdkChannel)
	mux.HandleFunc("/", viewexamples.Example)
	mux.HandleFunc("/og/", viewog.Values)
	mux.HandleFunc("/rog/", viewog.Base64)
	mux.HandleFunc("/rog-redirect/", viewog.Redirect)
	mux.HandleFunc(oauth.Path, oauth.Handle)
	mux.HandleFunc("/sleep/", httpdev.Sleep)

	handler = &httpstats.Handler{
		Name:    "web",
		Handler: mux,
		Stats:   service.Stats,
	}
	handler = &appdata.Handler{
		Handler: handler,
		Secret:  fbapp.Default.SecretByte(),
	}
	handler = httpgzip.NewHandler(handler)
	return handler
}
