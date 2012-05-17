// Command rell implements the main webserver application for Rell.
package main

import (
	"flag"
	"github.com/nshah/go.browserify"
	"github.com/nshah/go.fbapp"
	"github.com/nshah/go.flag.pkgpath"
	"github.com/nshah/go.flagconfig"
	"github.com/nshah/go.httpstats"
	"github.com/nshah/go.signedrequest/appdata"
	"github.com/nshah/rell/context/viewcontext"
	"github.com/nshah/rell/examples/viewexamples"
	"github.com/nshah/rell/og/viewog"
	"log"
	"net/http"
	"path/filepath"
)

var (
	serverAddress = flag.String(
		"rell.address",
		":43600",
		"Server address to bind to.")
	publicDir = pkgpath.Dir(
		"rell.public",
		"github.com/nshah/rell/public",
		"The directory to serve static files from.")
)

func main() {
	const public = "/public/"
	flag.Parse()
	flagconfig.Parse()

	mux := http.NewServeMux()
	staticFile(mux, "/favicon.ico")
	mux.Handle(public,
		http.StripPrefix(public, http.FileServer(http.Dir(*publicDir))))
	mux.HandleFunc(browserify.Path, browserify.Handle)
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

	var handler http.Handler
	handler = httpstats.NewHandler("web", mux)
	handler = &appdata.Handler{
		Handler: handler,
		Secret:  fbapp.Default.SecretByte(),
	}

	log.Println("Listening on ", *serverAddress)
	err := http.ListenAndServe(*serverAddress, handler)
	if err != nil {
		log.Fatalln("ListenAndServe: ", err)
	}
	log.Println("Exiting")
}

// binds a path to a single file
func staticFile(mux *http.ServeMux, name string) {
	abs := filepath.Join(*publicDir, name)
	mux.HandleFunc(name, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, abs)
	})
}
