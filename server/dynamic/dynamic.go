// Dynamic content rendering for the web side.
package dynamic

import (
	"embed"
	"html/template"
	"io"
	"io/fs"
	"time"

	"github.com/cceckman/reading-list/server/entry"
	"github.com/cceckman/reading-list/server/paths"
)

//go:embed *.html body/* menu/*
var templates embed.FS
var embeddedTemplates *template.Template

func init() {
	embeddedTemplates = loadTemplates(templates)
}

func loadTemplates(fs fs.FS) *template.Template {
	funcs := template.FuncMap{
		"maybeDate": maybeDate,
		"orToday":   orToday,
	}

	return template.Must(template.New("main.html").Funcs(funcs).ParseFS(fs, "*.html", "*/*.html"))
}

// Return a renderer that uses templates embedded in the binary.
func New() Renderer {
	return embeddedTemplates.Lookup
}

func NewFromFs(fs fs.FS) Renderer {
	return func(name string) *template.Template {
		// Re-load all templates on every lookup.
		// Yes, this is expensive; but it means we get "live" templates at every refresh.
		return loadTemplates(fs)
	}
}

// Renderer renders dynamic content for the site.
type Renderer getTemplate

// Indirection layer for "get template".
// This allows rewriting templates at runtime when operating in "local" mode.
type getTemplate func(name string) *template.Template

// Formatting function for entry dates.
func maybeDate(d time.Time) string {
	if d.IsZero() {
		return "—"
	} else {
		return d.Format(entry.DateFormat)
	}
}

// Alternative formatting function: default to "today"
func orToday(d time.Time) string {
	if d.IsZero() {
		return time.Now().Format(entry.DateFormat)
	} else {
		return d.Format(entry.DateFormat)
	}
}

// Fill for templates.
type fill struct {
	paths.Paths
	CurrentItem *entry.Entry
	ListItems   []*entry.Entry
}

// Render the "list" page to the provided writer, using the provided entries.
func (r Renderer) List(w io.Writer, paths paths.Paths, entries []*entry.Entry) error {
	return r("main.html").Execute(w, fill{Paths: paths, ListItems: entries})
}

// Render the "edit" page for the provided entry.
func (r Renderer) Edit(w io.Writer, paths paths.Paths, entry *entry.Entry) error {
	return r("main.html").Execute(w, fill{Paths: paths, CurrentItem: entry})
}
