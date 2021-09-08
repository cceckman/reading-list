package main

import (
	"log"
	"net/http"
)

func handleNew(w http.ResponseWriter, r *http.Request) {
	log.Printf("got share request: %+v", r)
	w.Write([]byte("share request logged"))
}

func main() {
	http.HandleFunc("/reading/admin/new", handleNew)

	http.Handle("/reading/admin/",
		http.StripPrefix("/reading/admin", http.FileServer(http.Dir(""))))

	log.Fatal(http.ListenAndServeTLS(":8080", "localhost.pem", "localhost-key.pem", nil))
}
