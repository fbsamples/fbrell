// Command rell implements the main webserver application for Rell.
package main

import (
	"flag"
	"github.com/nshah/go.basicssl"
	"github.com/nshah/go.browserify"
	"github.com/nshah/go.fbapp"
	"github.com/nshah/go.flag.pkgpath"
	"github.com/nshah/go.flagconfig"
	"github.com/nshah/go.httpstats"
	"github.com/nshah/go.signedrequest/appdata"
	"github.com/nshah/go.static"
	"github.com/nshah/go.viewvar"
	"github.com/nshah/rell/context/viewcontext"
	"github.com/nshah/rell/examples/viewexamples"
	"github.com/nshah/rell/og/viewog"
	"github.com/rcrowley/goagain"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"path/filepath"
)

const adminPath = "/admin/"

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
	flag.Parse()
	flagconfig.Parse()

	l, ppid, err := goagain.GetEnvs()
	if err != nil {
		addr, err := net.ResolveTCPAddr("tcp", *serverAddress)
		if err != nil {
			log.Fatal(err)
		}
		l, err = net.ListenTCP("tcp", addr)
		if err != nil {
			log.Fatal(err)
		}
		go serve(l)
	} else {
		log.Printf("Graceful handoff.")
		go serve(l)
		if err := goagain.KillParent(ppid); nil != err {
			log.Fatalf("Failed to kill parent: %s", err)
		}
	}
	if err := goagain.AwaitSignals(l); nil != err {
		log.Fatal(err)
	}
	log.Print("Exiting.")
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
	mux.HandleFunc(adminPath+"debug/pprof/", pprof.Index)
	mux.HandleFunc(adminPath+"debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc(adminPath+"debug/pprof/profile", pprof.Profile)
	mux.HandleFunc(adminPath+"debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc(adminPath+"vars/", viewvar.Json)
	return basicssl.Handler(mux)
}

func mainHandler() (handler http.Handler) {
	const public = "/public/"

	mux := http.NewServeMux()

	mux.Handle(adminPath, adminHandler())

	static.SetDir(*publicDir)
	mux.HandleFunc(static.Path, static.Handle)

	staticFile(mux, "/favicon.ico")
	staticFile(mux, "/f8.jpg")
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

	handler = httpstats.NewHandler("web", mux)
	handler = &appdata.Handler{
		Handler: handler,
		Secret:  fbapp.Default.SecretByte(),
	}
	return handler
}

func serve(l *net.TCPListener) {
	log.Println("Serving ", l.Addr())
	err := http.Serve(l, mainHandler())
	if err != nil {
		log.Fatalln("serve: ", err)
	}
}
