package entry

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"
	"time"
	"unicode"

	"github.com/gohugoio/hugo/parser/metadecoders"
	"github.com/gohugoio/hugo/parser/pageparser"
	yaml "gopkg.in/yaml.v3"
)

const entryKey string = "reading-list"
const DateFormat = "2006-01-02"

// Customize unmarshalling, so we can use short dates.
type Date struct {
	time.Time
}

// Lax date parsing: Allow extended format or date-only.
func ParseDate(s string) (Date, error) {
	var d Date
	var rawTime time.Time
	var err error
	if rawTime, err = time.Parse(time.RFC3339, s); err == nil {
		d.Time = rawTime
	} else if rawTime, err = time.Parse(DateFormat, s); err == nil {
		d.Time = rawTime
	} else if err != nil {
		return d, err
	}
	return d, nil
}

func (d *Date) UnmarshalYAML(value *yaml.Node) error {
	if dd, err := ParseDate(value.Value); err != nil {
		return err
	} else {

		*d = dd
		return nil
	}
}

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
	Added Date `yaml:",omitempty"`

	// When this entry was moved from "in the queue" to "read".
	Read Date `yaml:",omitempty"`
	// When commentary on this entry was made available.
	Reviewed Date `yaml:",omitempty"`

	// Discovery data: how did I come across this item?
	// This may be rendered as "found via..."
	Discovery *Source `yaml:",omitempty"`

	// Original content as read from storage. This allows the entire entry to be re-serialized.
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
// ID is not readable from the file itself; it is derived from e.g. the filename.
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
func (e *Entry) WriteTo(w io.Writer) (int64, error) {
	return 0, fmt.Errorf("unimplemented")
}

func genId(title string) string {
	id := title
	id = strings.TrimSpace(id)
	// Slugify: make lowercase, and keep only ASCII alnum;
	// everything else is a space.
	id = strings.Map(func(r rune) rune {
		r = unicode.ToLower(r)
		keep := (unicode.IsDigit(r) || unicode.IsLetter(r)) && r <= unicode.MaxASCII
		if keep {
			return r
		} else {
			return ' '
		}
	}, id)
	// Use `fields` to slugify.
	return strings.Join(strings.Fields(id), "-")
}

// Makes a source, if the provided text and URL are nonemty.
// Otherwise returns nil.
func makeSource(text, uri string) *Source {
	if text == "" && uri == "" {
		return nil
	}
	s := &Source{
		Text: text,
		Uri:  uri,
	}
	if s.Text == "" {
		if e, err := url.Parse(uri); err == nil {
			s.Text = e.Host
		}
	}
	return s
}

// Constructs an Entry from the given "save" or "share" form.
func FromForm(form url.Values) (*Entry, error) {
	log.Printf("parsing form: %+v", form)

	title := form.Get("title")
	id := form.Get("id")
	// If this is a new entry, we need to generate an ID.
	if id == "" {
		id = genId(title)
	}

	// The source URL may come from many places:
	// The save form and share form go to the same place:
	u := form.Get("source-url")
	if u == "" {
		// On Android, the share form's `text` field
		if url, err := url.Parse(form.Get("text")); err == nil {
			u = url.String()
		}
	}

	source := form.Get("source")
	// If we're coming from the share API, we may not have a source.
	// Use the hostname of the URL if we don't have something better.
	if u, err := url.Parse(u); err == nil && source == "" {
		source = u.Host
	}

	e := &Entry{
		Id:    id,
		Title: title,
		Source: Source{
			Text: source,
			Uri:  u,
		},
		Summary:   form.Get("summary"),
		Added:     Date{time.Now()},
		Discovery: makeSource(form.Get("discovery"), form.Get("discovery-url")),
		Author:    makeSource(form.Get("author"), form.Get("author-url")),
	}

	var err error
	// If this is an edit rather than a share:
	if form.Has("added") {
		e.Added, err = ParseDate(form.Get("added"))
		if err != nil {
			return nil, fmt.Errorf("invalid added date: %w", err)
		}
	}
	if form.Has("read") {
		e.Read, err = ParseDate(form.Get("read"))
		if err != nil {
			return nil, fmt.Errorf("invalid read date: %w", err)
		}
	}
	if form.Has("reviewed") {
		e.Reviewed, err = ParseDate(form.Get("reviewed"))
		if err != nil {
			return nil, fmt.Errorf("invalid reviewed date: %w", err)
		}
	}

	return e, nil
}
