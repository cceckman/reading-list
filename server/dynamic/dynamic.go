// Dynamic content rendering for the web side.
package dynamic

import (
	"html/template"
	"io"

	"embed"

	"github.com/cceckman/reading-list/server/entry"
)

//go:embed *.html body/* menu/*
var templates embed.FS

var parsedTemplates *template.Template = template.Must(template.ParseFS(templates, "*.html", "*/*.html"))

// URL paths to use when rendering templates.
// TODO: Use this more broadly.
type Paths interface {
	Edit() string
	Save() string
}

// Fill for templates.
type fill struct {
	Paths
	CurrentItem *entry.Entry
	ListItems   []entry.Entry
}

// Render the "list" page to the provided writer, using the provided entries.
func List(w io.Writer, paths Paths, entries []entry.Entry) error {
	return parsedTemplates.ExecuteTemplate(w, "main.html", fill{Paths: paths, ListItems: entries})
}

// Render the "edit" page for the provided entry.
func Edit(w io.Writer, paths Paths, entry *entry.Entry) error {
	return parsedTemplates.ExecuteTemplate(w, "main.html", fill{Paths: paths, CurrentItem: entry})
}
