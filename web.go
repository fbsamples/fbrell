// Command rell implements the main webserver application for Rell.
package main

import (
	"flag"
	"github.com/nshah/go.browserify"
	"github.com/nshah/go.fbapp"
	"github.com/nshah/go.finmonitor"
	"github.com/nshah/go.flag.pkgpath"
	"github.com/nshah/go.flagconfig"
	"github.com/nshah/go.httpstats"
	"github.com/nshah/go.pidfile"
	"github.com/nshah/go.signedrequest/appdata"
	"github.com/nshah/go.static"
	"github.com/nshah/go.viewvar"
	"github.com/nshah/rell/context/viewcontext"
	"github.com/nshah/rell/examples/viewexamples"
	"github.com/nshah/rell/oauth"
	"github.com/nshah/rell/og/viewog"
	"github.com/rcrowley/goagain"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
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
	adminToken = flag.String(
		"rell.admin.token",
		"",
		"Admin token.")
)

func main() {
	flag.Parse()
	flagconfig.Parse()
	pidfile.Write()
	finMonitor := finmonitor.New(mainHandler())

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
		log.Printf("Serving %s with pid %d.", l.Addr(), os.Getpid())
		go serve(l, finMonitor)
	} else {
		log.Printf(
			"Graceful handoff of %s with new pid %d and old pid %d.",
			l.Addr(), os.Getpid(), os.Getppid())
		go serve(l, finMonitor)
		if err := goagain.KillParent(ppid); nil != err {
			log.Fatalf("Failed to kill parent: %s", err)
		}
	}
	if err := goagain.AwaitSignals(l); nil != err {
		log.Fatal(err)
	}
	finMonitor.Notify <- true
	<-finMonitor.Done
	log.Printf("Exiting pid %d.", os.Getpid())
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
	mux.HandleFunc(adminPath+*adminToken+"/debug/pprof/", pprof.Index)
	mux.HandleFunc(adminPath+*adminToken+"/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc(adminPath+*adminToken+"/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc(adminPath+*adminToken+"/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc(adminPath+*adminToken+"/vars/", viewvar.Json)
	return mux
}

func mainHandler() (handler http.Handler) {
	const public = "/public/"

	mux := http.NewServeMux()

	mux.Handle(adminPath, adminHandler())

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

	handler = httpstats.NewHandler("web", mux)
	handler = &appdata.Handler{
		Handler: handler,
		Secret:  fbapp.Default.SecretByte(),
	}
	return handler
}

func serve(l *net.TCPListener, handler http.Handler) {
	err := http.Serve(l, handler)
	if err != nil {
		log.Fatalln("serve: ", err)
	}
}
