package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
)

var port = flag.Int("port", 8080, "Port for the HTTP server to listen on")

const SERVER_ENV = "READING_LIST_SERVER"

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

func getTarget() (*url.URL, error) {
	t := os.Getenv(SERVER_ENV)
	if t == "" {
		t = "http://localhost:8080/entries"
	}
	return url.Parse(t)
}

func handleAdd(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	log.Printf("got share request: with query params: %+v", queryParams)

	newUrl, _ := getTarget()
	newUrl.RawQuery = queryParams.Encode()

	// Use 307 Temporary Redirect to preserve the method,
	// and so the client knows to re-check this page every time.
	http.Redirect(w, r, newUrl.String(), http.StatusTemporaryRedirect)
}

func main() {
	flag.Parse()

	if _, err := getTarget(); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/add", handleAdd)
	http.Handle("/", http.FileServer(http.Dir("static")))

	listen := fmt.Sprintf(":%d", *port)

	if hasKey("localhost.pem", "localhost-key.pem") {
		log.Print("Listening for HTTPS connection...")
		log.Fatal(http.ListenAndServeTLS(listen, "localhost.pem", "localhost-key.pem", nil))
	} else {
		log.Print("Listening for HTTPS connection...")
		log.Fatal(http.ListenAndServe(listen, nil))
	}
}
