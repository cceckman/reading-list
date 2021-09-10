package main

import (
	"log"
	"net/http"
	"os"

	"github.com/cceckman/reading-list/server"
)

// TODO: Temporary storage location, in the current directory.
const storage = "test-storage"
const listen = "[::1]:8080"

func main() {
	if err := os.MkdirAll(storage, 0755); err != nil {
		panic(err)
	}

	dfs := os.DirFS(storage)
	srv := server.New(dfs)

	log.Printf("Listening on %s", listen)
	http.ListenAndServe(listen, &srv)
}
