// Command rell implements the main webserver application for Rell.
package main

import (
	"flag"
	"github.com/nshah/go.browserify"
	"github.com/nshah/go.httpstats"
	"github.com/nshah/go.viewvar"
	"github.com/nshah/rell/context/viewcontext"
	"github.com/nshah/rell/examples/viewexamples"
	"github.com/nshah/rell/og/viewog"
	"go/build"
	"log"
	"net/http"
)

var (
	ServerAddress = flag.String(
		"rell.address",
		":43600",
		"Server address to bind to.")
)

// Static files are served from a directory within the package.
func getPublicDir() string {
	pkg, err := build.Import(
		"github.com/nshah/rell/public", "", build.FindOnly)
	if err != nil {
		log.Fatalf("Failed to find public directory: %s", err)
	}
	return pkg.Dir
}

func main() {
	const public = "/public/"
	flag.Parse()
	mux := http.NewServeMux()
	mux.Handle(public,
		http.StripPrefix(public, http.FileServer(http.Dir(getPublicDir()))))
	mux.HandleFunc(browserify.Path, browserify.Handle)
	mux.HandleFunc("/debug/var/", viewvar.Json)
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
	log.Println("Listening on ", *ServerAddress)
	err := http.ListenAndServe(*ServerAddress, httpstats.NewHandler("web", mux))
	if err != nil {
		log.Fatalln("ListenAndServe: ", err)
	}
}
