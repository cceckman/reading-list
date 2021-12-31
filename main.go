package main

import (
	"crypto/tls"
	"flag"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/cceckman/reading-list/static"
	"tailscale.com/client/tailscale"
	"tailscale.com/tsnet"
)

var (
	addr       = flag.String("addr", ":443", "Port or address:port to listen on")
	allowLocal = flag.Bool("allowLocal", false, "Allow serving from the local static/ directory rather than embedded content. Development only.")
	useTsNet   = flag.Bool("tsnet", true, "Connect directly to Tailscale via tsnet")
)

func main() {
	flag.Parse()

	var files fs.FS
	if *allowLocal {
		log.Print("Serving from local directories")
		// Serve from local filesystem.
		files = os.DirFS("static")
	} else {
		log.Print("Serving from embedded files")
		files = static.Files
	}
	// Root handler: serve from the filesystem.
	fileHandler := http.FileServer(http.FS(files))

	var ln net.Listener
	var err error
	if *useTsNet {
		s := new(tsnet.Server)
		ln, err = s.Listen("tcp", *addr)
	} else {
		ln, err = net.Listen("tcp", *addr)
	}
	if err != nil {
		log.Fatal(err)
	}

	tls := tls.NewListener(ln, &tls.Config{
		GetCertificate: tailscale.GetCertificate,
	})

	log.Fatal(http.Serve(tls, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if *useTsNet {
			// Check Tailscale authentication
			who, err := tailscale.WhoIs(r.Context(), r.RemoteAddr)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			log.Printf("Tailscale-authenticated request from %s at %s", who.UserProfile.LoginName, who.Node.Name)
		}
		fileHandler.ServeHTTP(w, r)
	})))
}
