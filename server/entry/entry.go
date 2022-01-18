package entry

import (
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/gohugoio/hugo/parser/metadecoders"
	"github.com/gohugoio/hugo/parser/pageparser"
	yaml "gopkg.in/yaml.v3"
)

const entryKey string = "reading-list"

// Metadata for a reading-list entry.
type Entry struct {
	// ID of the entry; e.g. a slug.
	// Note that this is provided when constructing the entry.
	Id string `yaml:"-"`

	// Title of the entry.
	// Part of Hugo's metadata.
	Title string `yaml:"-"`

	// Summary of the entry.
	// Part of Hugo's metadata.
	Summary string `yaml:"-"`

	// Source: how can this item be found on the internet?
	Source Source `yaml:"source,omitempty"`

	// Who wrote this work, and where can they be found?
	Author *Source `yaml:",omitempty"`

	// When was this item added to the reading list?
	Added time.Time `yaml:",omitempty"`

	// When this entry was moved from "in the queue" to "read".
	Read time.Time `yaml:",omitempty"`
	// When commentary on this entry was made available.
	Reviewed time.Time `yaml:",omitempty"`

	// Discovery data: how did I come across this item?
	// This may be rendered as "found via..."
	Discovery *Source `yaml:",omitempty"`

	// Original content as read out. This allows the entire entry to be re-serialized.
	original pageparser.ContentFrontMatter `yaml:"-"`
}

// Discovery / link metadata.
type Source struct {
	Text string
	Uri  string `yaml:"uri,omitempty"`
}

func (s *Source) UrlString() string {
	if u, err := url.Parse(s.Uri); err == nil {
		return u.String()
	}
	return ""
}

// Get a string from a string-to-anything map.
func getStringProperty(key string, m map[string]interface{}) (string, error) {
	if _, ok := m[key]; !ok {
		return "", fmt.Errorf("no key %s", key)
	}

	if v, ok := m[key].(string); ok {
		return v, nil
	}
	if v, ok := m[key].(fmt.Stringer); ok {
		return v.String(), nil
	}

	return "", fmt.Errorf("%s cannot be interpreted as a string", key)
}

// Read the front-matter from the input channel.
// ID is not readable from the
func Read(id string, r io.Reader) (*Entry, error) {
	cfm, err := pageparser.ParseFrontMatterAndContent(r)
	if err != nil {
		return nil, fmt.Errorf("could not get reading list entry from %s: %w", id, err)
	}
	if cfm.FrontMatterFormat != metadecoders.YAML {
		return nil, fmt.Errorf("reading list entry for %s is not in YAML format", id)
	}
	properties := cfm.FrontMatter

	var title string
	var summary string
	if title, err = getStringProperty("title", properties); err != nil {
		return nil, fmt.Errorf("could not get title for %s: %w", id, err)
	}
	if summary, err = getStringProperty("summary", properties); err != nil {
		// Allow summary to be empty.
		summary = ""
	}

	// We re-marshal and unmarshal into our own type.
	// Ick; but it works.
	var entryBytes []byte
	if entry, ok := properties[entryKey]; ok {
		if entryBytes, err = yaml.Marshal(entry); err != nil {
			return nil, fmt.Errorf("could not reencode reading list entry %s: %w", id, err)
		}
	}

	e := Entry{
		Id:      id,
		Title:   title,
		Summary: summary,
	}
	if err := yaml.Unmarshal(entryBytes, &e); len(entryBytes) != 0 && err != nil {
		return nil, fmt.Errorf("error decoding reading list entry %s: %w", id, err)
	}
	e.Id = id
	e.Title = title
	e.original = cfm

	return &e, nil
}

// Marshal the item back to a writer, e.g. a file.
func (e *Entry) WriteTo(w io.Writer) (int, error) {
	return 0, fmt.Errorf("unimplemented")
}
