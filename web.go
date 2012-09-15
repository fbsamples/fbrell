// Command rell implements the main webserver application for Rell.
package main

import (
	"flag"
	"github.com/daaku/go.browserify"
	"github.com/daaku/go.fbapp"
	"github.com/daaku/go.flag.pkgpath"
	"github.com/daaku/go.flagconfig"
	"github.com/daaku/go.flagenv"
	"github.com/daaku/go.grace/gracehttp"
	"github.com/daaku/go.httpdev"
	"github.com/daaku/go.httpstats"
	"github.com/daaku/go.pidfile"
	"github.com/daaku/go.signedrequest/appdata"
	"github.com/daaku/go.static"
	"github.com/daaku/go.stats"
	"github.com/daaku/go.stats/stathatbackend"
	"github.com/daaku/go.viewvar"
	"github.com/daaku/rell/context/viewcontext"
	"github.com/daaku/rell/examples/viewexamples"
	"github.com/daaku/rell/oauth"
	"github.com/daaku/rell/og/viewog"
	"log"
	"net/http"
	"net/http/pprof"
	"path/filepath"
	"runtime"
)

var (
	mainAddress = flag.String(
		"rell.address",
		":43600",
		"Server address to bind to.")
	adminAddress = flag.String(
		"rell.admin.address",
		":43601",
		"Admin http server address.")
	publicDir = pkgpath.Dir(
		"rell.public",
		"github.com/daaku/rell/public",
		"The directory to serve static files from.")
	goMaxProcs = flag.Int(
		"rell.gomaxprocs",
		runtime.NumCPU(),
		"Maximum processes to use.")
	ezkey = flag.String("rell.stats.key", "", "The stathat ezkey.")
)

func main() {
	runtime.GOMAXPROCS(*goMaxProcs)
	flag.Parse()
	flagconfig.Parse()
	flagenv.Parse()
	pidfile.Write()
	stats.SetBackend(stathatbackend.EZKey(*ezkey))
	err := gracehttp.Serve(
		gracehttp.Handler{*mainAddress, mainHandler()},
		gracehttp.Handler{*adminAddress, adminHandler()},
	)
	if err != nil {
		log.Fatal(err)
	}
}

// binds a path to a single file
func staticFile(mux *http.ServeMux, name string) {
	abs := filepath.Join(*publicDir, name)
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

	static.SetDir(*publicDir)
	mux.HandleFunc(static.Path, static.Handle)

	staticFile(mux, "/favicon.ico")
	staticFile(mux, "/f8.jpg")
	staticFile(mux, "/robots.txt")
	mux.Handle(public,
		http.StripPrefix(public, http.FileServer(http.Dir(*publicDir))))

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

	handler = httpstats.NewHandler("web", mux)
	handler = &appdata.Handler{
		Handler: handler,
		Secret:  fbapp.Default.SecretByte(),
	}
	return handler
}
