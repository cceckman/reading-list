package main

import (
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
)

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
		t = "https://cceckman.com"
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
	if _, err := getTarget(); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("add", handleAdd)
	http.Handle("/", http.FileServer(http.Dir("static")))

	if hasKey("localhost.pem", "localhost-key.pem") {
		log.Print("Listening for HTTPS connection...")
		log.Fatal(http.ListenAndServeTLS(":8080", "localhost.pem", "localhost-key.pem", nil))
	} else {
		log.Print("Listening for HTTPS connection...")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}
}
