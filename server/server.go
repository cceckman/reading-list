package server

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/url"

	"github.com/cceckman/reading-list/server/dynamic"
	"github.com/cceckman/reading-list/server/entry"
	"github.com/cceckman/reading-list/server/paths"
	"github.com/cceckman/reading-list/server/static"
)

// Interface for managing entries.
type EntryManager interface {
	// Create(entry.Entry) error
	Read(id string) (*entry.Entry, error)
	Update(*entry.Entry) error
	List(limit int) ([]*entry.Entry, error)
}

// Return a server for the entry manager, rendering based on embedded templates.
func New(paths paths.Paths, em EntryManager) *Server {
	s := &Server{
		paths:   paths,
		manager: em,
		static:  http.FileServer(http.FS(static.Files)),
		dynamic: dynamic.New(),
		mux:     http.NewServeMux(),
	}
	s.setupRouter()
	return s
}

// Return a server for the entry manager, rendering from templates live on the filesystem.
func NewFs(paths paths.Paths, em EntryManager, static fs.FS, templates fs.FS) *Server {
	s := &Server{
		paths:   paths,
		manager: em,
		static:  http.FileServer(http.FS(static)),
		dynamic: dynamic.NewFromFs(templates),
		mux:     http.NewServeMux(),
	}
	s.setupRouter()
	return s
}

// Server serves an app managing reading-list entries.
type Server struct {
	paths   paths.Paths
	manager EntryManager
	static  http.Handler
	dynamic dynamic.Renderer

	mux *http.ServeMux
}

func (s *Server) setupRouter() {
	s.mux.Handle(paths.Default.List(), http.HandlerFunc(s.serveList))
	s.mux.Handle(paths.Default.Edit(), http.HandlerFunc(s.serveEdit))
	s.mux.Handle(paths.Default.Save(), http.HandlerFunc(s.serveSave))
	s.mux.Handle(paths.Default.Share(), http.HandlerFunc(s.serveShare))
	s.mux.Handle("/", http.HandlerFunc(s.serveDefault))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if f, ok := (w.(http.Flusher)); ok {
		defer f.Flush()
	}
	s.mux.ServeHTTP(w, r)
}

func (s *Server) serveList(w http.ResponseWriter, r *http.Request) {
	// TODO: Don't have an arbitrary queue length; do better filtering in other dimensions.
	items, err := s.manager.List(100)
	if err != nil {
		log.Printf("error in serving list request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error in serving list request: %v", err)
		return
	}
	if err := s.dynamic.List(w, s.paths, items); err != nil {
		log.Printf("error rendering list template: %v", err)
	}
}

func (s *Server) serveEdit(w http.ResponseWriter, r *http.Request) {
	// Do we have an ID?
	id := r.URL.Query().Get("id")
	var e *entry.Entry
	var err error
	if id != "" {
		e, err = s.manager.Read(id)
	} else {
		e = &entry.Entry{}
	}
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := s.dynamic.Edit(w, s.paths, e); err != nil {
		log.Printf("error rendering list template: %v", err)
	}

}

func (s *Server) serveSave(w http.ResponseWriter, r *http.Request) {
	ent, err := s.serveUpdate(w, r)
	if err != nil {
		log.Printf("error serving save request: %v", err)
		return
	}

	log.Print("manager saved entry")
	var donePath url.URL
	{
		done := make(url.Values)
		done.Add("done", ent.Id)
		donePath.Path = s.paths.List()
		donePath.RawQuery = done.Encode()
	}

	w.Header().Add("Location", donePath.String())
	w.WriteHeader(http.StatusSeeOther)
}

func (s *Server) serveShare(w http.ResponseWriter, r *http.Request) {
	ent, err := s.serveUpdate(w, r)
	if err != nil {
		log.Printf("error serving share request: %v", err)
		return
	}

	// Redirect to edit location one the share is complete.
	log.Print("manager saved new entry")
	var editPath url.URL
	{
		edit := make(url.Values)
		edit.Add("id", ent.Id)
		editPath.Path = s.paths.Edit()
		editPath.RawQuery = edit.Encode()
	}

	w.Header().Add("Location", editPath.String())
	w.WriteHeader(http.StatusSeeOther)
}

func (s *Server) serveUpdate(w http.ResponseWriter, r *http.Request) (ent *entry.Entry, finerr error) {
	log.Printf("invoking update handler with request: %+v", r)
	defer func() {
		if finerr != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Failed to update: %v", finerr)
		}
	}()

	if err := r.ParseForm(); err != nil {
		finerr = fmt.Errorf("error in parsing form data: %w", err)
		return
	}

	e, err := entry.FromForm(r.Form)
	if err != nil {
		finerr = fmt.Errorf("error in interpreting form as entry: %w", err)
		return
	}

	log.Printf("invoking update handler with entry: %+v", e)

	if err := s.manager.Update(e); err != nil {
		finerr = fmt.Errorf("error in updating entry: %w", err)
		return
	}

	return e, nil
}

func (s *Server) serveDefault(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "" || r.URL.Path == "/" {
		s.serveList(w, r)
		return
	}
	s.static.ServeHTTP(w, r)
}
