// Dynamic content rendering for the web side.
package dynamic

import (
	"html/template"
	"io"

	"embed"

	"github.com/cceckman/reading-list/server/entry"
	"github.com/cceckman/reading-list/server/paths"
)

//go:embed *.html body/* menu/*
var templates embed.FS

var parsedTemplates *template.Template = template.Must(template.ParseFS(templates, "*.html", "*/*.html"))

// Fill for templates.
type fill struct {
	paths.Paths
	CurrentItem *entry.Entry
	ListItems   []entry.Entry
}

// Render the "list" page to the provided writer, using the provided entries.
func List(w io.Writer, paths paths.Paths, entries []entry.Entry) error {
	return parsedTemplates.ExecuteTemplate(w, "main.html", fill{Paths: paths, ListItems: entries})
}

// Render the "edit" page for the provided entry.
func Edit(w io.Writer, paths paths.Paths, entry *entry.Entry) error {
	return parsedTemplates.ExecuteTemplate(w, "main.html", fill{Paths: paths, CurrentItem: entry})
}
