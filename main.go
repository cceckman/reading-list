package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/cceckman/reading-list/server"
	"github.com/spf13/afero"
)

var (
	storage = flag.String("storage", "/var/lib/reading-list", "path to store pending items")
	listen  = flag.String("listen", "[::]:8080", "Port or address:port to listen on")
)

func main() {
	flag.Parse()

	if err := os.MkdirAll(*storage, 0755); err != nil {
		panic(err)
	}

	dfs := afero.NewBasePathFs(afero.NewOsFs(), *storage)
	srv := server.New(dfs)

	log.Printf("Listening on %s", *listen)
	http.ListenAndServe(*listen, &srv)
}
