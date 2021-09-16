package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
)

var port = flag.Int("port", 8080, "Port for the HTTP server to listen on")

func hasKey(public, private string) bool {
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}
	if _, err := fs.Stat(os.DirFS(cwd), public); err != nil {
		return false
	}
	if _, err := fs.Stat(os.DirFS(cwd), private); err != nil {
		return false
	}
	return true
}

func main() {
	flag.Parse()

	http.Handle("/", http.FileServer(http.Dir("static")))

	listen := fmt.Sprintf(":%d", *port)

	if hasKey("localhost.pem", "localhost-key.pem") {
		log.Printf("Listening for HTTPS connection at %s", listen)
		log.Fatal(http.ListenAndServeTLS(listen, "localhost.pem", "localhost-key.pem", nil))
	} else {
		log.Printf("Listening for HTTP connection... at %s", listen)
		log.Fatal(http.ListenAndServe(listen, nil))
	}
}
