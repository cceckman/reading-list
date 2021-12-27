package main

import (
	"crypto/tls"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/cceckman/reading-list/static"
	"tailscale.com/client/tailscale"
)

var (
	listen     = flag.String("listen", "[::]:8080", "Port or address:port to listen on")
	allowLocal = flag.Bool("allowLocal", false, "Allow serving from the local static/ directory rather than embedded content. Development only.")
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

	server := http.Server{
		Addr:    *listen,
		Handler: fileHandler,
		TLSConfig: &tls.Config{
			GetCertificate: tailscale.GetCertificate,
		},
	}
	log.Print("Starting to run and listen at ", *listen)
	log.Fatal(server.ListenAndServeTLS("", ""))
}
