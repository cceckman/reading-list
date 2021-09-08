package main

import (
	"log"
	"net/http"
	"net/url"
)

func handleNew(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	log.Printf("got share request: with query params: %+v", queryParams)

	newUrl := &url.URL{
		Scheme:   "https",
		Host:     "cceckman.com",
		Path:     "reading-list",
		RawQuery: queryParams.Encode(),
	}
	// Use 307 Temporary Redirect to preserve the method,
	// and so the client knows to re-check this page every time.
	http.Redirect(w, r, newUrl.String(), http.StatusTemporaryRedirect)
}

func main() {
	http.HandleFunc("/reading/admin/new", handleNew)

	http.Handle("/reading/admin/",
		http.StripPrefix("/reading/admin", http.FileServer(http.Dir(""))))

	log.Fatal(http.ListenAndServeTLS(":8080", "localhost.pem", "localhost-key.pem", nil))
}
