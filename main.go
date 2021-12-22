package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"

	"github.com/cceckman/reading-list/static"
	"tailscale.com/client/tailscale"
)

var (
	listen = flag.String("listen", "[::]:8080", "Port or address:port to listen on")
)

func main() {
	flag.Parse()

	// Root handler: serve from the filesystem.
	fileHandler := http.FileServer(http.FS(static.Files))

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
