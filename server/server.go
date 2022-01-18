package server

import (
	"io/fs"
	"net/http"

	"github.com/cceckman/reading-list/server/dynamic"
	"github.com/cceckman/reading-list/server/entry"
	"github.com/cceckman/reading-list/server/paths"
	"github.com/cceckman/reading-list/server/static"
)

// Return a new server that provides content from embedded data.
func New(paths paths.Paths) *Server {
	s := &Server{
		paths:   paths,
		static:  http.FileServer(http.FS(static.Files)),
		dynamic: dynamic.New(),
	}
	s.setupRouter()
	return s
}

// Return a new server that provides content from the filesystem.
func NewFs(paths paths.Paths, static fs.FS, templates fs.FS) *Server {
	s := &Server{
		paths:   paths,
		static:  http.FileServer(http.FS(static)),
		dynamic: dynamic.NewFromFs(templates),
	}
	s.setupRouter()
	return s
}

// Server serves an app managing reading-list entries.
type Server struct {
	paths   paths.Paths
	static  http.Handler
	dynamic dynamic.Renderer

	mux *http.ServeMux
}

func (s *Server) setupRouter() {
	s.mux.Handle(paths.Default.List(), http.HandlerFunc(s.serveList))
	s.mux.Handle("/", http.HandlerFunc(s.serveDefault))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) serveList(w http.ResponseWriter, r *http.Request) {
	s.dynamic.List(w, s.paths, []entry.Entry{})
}

func (s *Server) serveDefault(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "" || r.URL.Path == "/" {
		s.serveList(w, r)
	}
	s.static.ServeHTTP(w, r)
}
