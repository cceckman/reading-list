package main

import (
	"log"
	"net/http"
)

func main() {
	http.Handle("/reading/admin/", http.StripPrefix("/reading/admin", http.FileServer(http.Dir(""))))
	log.Fatal(http.ListenAndServeTLS(":8080", "localhost.pem", "localhost-key.pem", nil))
}
