package server

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/cceckman/reading-list/entry"
)

const (
	textFormKey  = "text"
	titleFormKey = "title"
	urlFormKey   = "url"
)

// Server allows managing reading-list entries in local storage via a CRUD layer.
type Server struct {
	entry.EntryManager
}

func New(dir fs.FS) Server {
	return Server{
		EntryManager: entry.NewManager(dir),
	}
}

// Allow convenient return-an-error
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := s.serve(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// URL mapping: REST-ful, even if we're showing it to a user
// POST /add to create
// GET /entry/<key> to read - eventually, show edit UI
// TODO: PUT /entry/<key> to send the update
func (s *Server) serve(w http.ResponseWriter, r *http.Request) error {
	log.Printf("request for %s", r.URL)
	if r.URL.Path == "/entries" {
		return s.serveAdd(w, r)
	}
	dir, key := path.Split(r.URL.Path)
	if dir == "/entries" {
		switch r.Method {
		case http.MethodGet:
			return s.serveGet(w, key)
		default:
			w.Header().Add("Allow", "GET")
			// TODO: support updates
			// w.Header().Add("Allow", "PUT")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return nil
		}
	}

	w.WriteHeader(http.StatusNotFound)
	return nil
}

func (s *Server) serveAdd(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		w.Header().Add("Allow", http.MethodPost)
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("use the POST method to add entries\n"))
		return nil
	}

	// On Chrome (OS), https://web.dev/web-share-target/ fills "text" and "url"-
	// but not "title".
	title := r.FormValue(titleFormKey)
	if title == "" {
		title = r.FormValue(textFormKey)
	}
	url := r.FormValue(urlFormKey)
	key, err := s.EntryManager.Create(title, url)
	if err != nil {
		return err
	}

	w.Header().Add("Location", path.Join("/entries", key))
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "created entry: %s\n", key)
	return nil
}

func (s *Server) serveGet(w http.ResponseWriter, key string) error {
	entry, err := s.EntryManager.Read(key)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			w.WriteHeader(http.StatusNotFound)
			return nil
		}
		return err
	}

	// TODO: Print this as an "edit" screen.
	fmt.Fprintf(w, "found entry: %+v", entry)
	return nil
}
