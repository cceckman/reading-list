package entry

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/gohugoio/hugo/parser/metadecoders"
	"github.com/gohugoio/hugo/parser/pageparser"
	yaml "gopkg.in/yaml.v3"
)

const entryKey string = "reading-list"

// ISO 8601 date format.
const DateFormat = "2006-01-02"

// Regex that matches valid IDs: a lowercase letter, followed by lowercase letters, numbers, and hyphens.
var idMatch = regexp.MustCompile("[a-z][a-z0-9-]+")

// Customize unmarshalling, so we can use short dates.
type Date struct {
	time.Time
}

// Sorts a list of *Entrys in FIFO order: oldest unread item first.
func FifoSort(list []*Entry) {
	sort.Slice(list, func(i, j int) bool {
		if list[i].Read.IsZero() && list[j].Read.IsZero() {
			if list[i].Added == list[j].Added {
				// Same date: sort by name.
				return list[i].Id < list[j].Id
			}
			// Later date is lower.
			return list[j].Added.After(list[i].Added.Time)
		}
		if list[i].Read.IsZero() && !list[j].Read.IsZero() {
			// Unread goes first.
			return true
		}
		// Both are read.
		if list[i].Read == list[j].Read {
			// Same date: sort by name.
			return list[i].Id < list[j].Id
		}
		// Later date is lower.
		return list[j].Read.After(list[i].Read.Time)
	})
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
	Author *Source `yaml:"author,omitempty"`

	// When was this item added to the reading list?
	Added Date `yaml:"added,omitempty"`

	// When this entry was moved from "in the queue" to "read".
	Read Date `yaml:"read,omitempty"`
	// When commentary on this entry was made available.
	Reviewed Date `yaml:"reviewed,omitempty"`

	// Discovery data: how did I come across this item?
	// This may be rendered as "found via..."
	Discovery *Source `yaml:"discovery,omitempty"`

	// Original content as read from storage. This allows the entire entry to be re-serialized.
	original pageparser.ContentFrontMatter `yaml:"-"`
}

// Check that the ID for the entry is valid.
func (e *Entry) ValidID() error {
	if idMatch.Match([]byte(e.Id)) {
		return nil
	} else {
		return fmt.Errorf("invalid ID: %q", e.Id)
	}
}

// Get the contents of the underlying entry.
func (e *Entry) Content() string {
	return string(e.original.Content)
}

// Discovery / link metadata.
type Source struct {
	Text string
	Uri  string `yaml:"uri,omitempty"`
}

// If the URI is a URL, return it; otherwise, return an empty string.
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
func Read(id string, r io.Reader) (Entry, error) {
	cfm, err := pageparser.ParseFrontMatterAndContent(r)
	if err != nil {
		return Entry{}, fmt.Errorf("could not get reading list entry from %s: %w", id, err)
	}
	if cfm.FrontMatterFormat != metadecoders.YAML {
		return Entry{}, fmt.Errorf("reading list entry for %s is not in YAML format", id)
	}
	properties := cfm.FrontMatter

	var title string
	var summary string
	if title, err = getStringProperty("title", properties); err != nil {
		return Entry{}, fmt.Errorf("could not get title for %s: %w", id, err)
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
			return Entry{}, fmt.Errorf("could not reencode reading list entry %s: %w", id, err)
		}
	}

	e := Entry{
		Id:      id,
		Title:   title,
		Summary: summary,
	}
	if err := yaml.Unmarshal(entryBytes, &e); len(entryBytes) != 0 && err != nil {
		return Entry{}, fmt.Errorf("error decoding reading list entry %s: %w", id, err)
	}
	e.Id = id
	e.Title = title
	e.original = cfm

	return e, nil
}

// Marshal the item back to a writer, e.g. a file.
func (e *Entry) WriteTo(w io.Writer) (count int64, err error) {
	// Create a new front-matter dictionary for writing.
	// This allows us to avoid a circular reference:
	//   new front matter --> Entry --> old front matter
	// rather than
	//   old front matter --> Entry --> old front matter
	fm := make(map[string]interface{}, len(e.original.FrontMatter))
	for k, v := range e.original.FrontMatter {
		fm[k] = v
	}
	fm[entryKey] = e
	// Items that we take from the Hugo info overwrite those keys.
	fm["title"] = e.Title
	fm["summary"] = e.Summary
	if e.Title == "" {
		return 0, fmt.Errorf("must have title for entry %q", e.Id)
	}

	// YAML start/end delimiter.
	const yamlDelim = "---\n"
	yamlBlock := bytes.NewBuffer(nil)
	enc := yaml.NewEncoder(yamlBlock)
	if err := enc.Encode(fm); err != nil {
		return 0, err
	}

	var n int
	var n64 int64

	n, err = io.WriteString(w, yamlDelim)
	count += int64(n)
	if err != nil {
		return
	}

	n64, err = yamlBlock.WriteTo(w)
	count += n64
	if err != nil {
		return
	}

	n, err = io.WriteString(w, yamlDelim)
	count += int64(n)
	if err != nil {
		return
	}

	content := bytes.NewBuffer(e.original.Content)
	n64, err = content.WriteTo(w)
	count += n64
	return
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
func FromForm(form url.Values) (Entry, error) {
	// log.Printf("parsing form: %+v", form)

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

	e := Entry{
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
	if form.Has("added") && form.Get("added") != "" {
		e.Added, err = ParseDate(form.Get("added"))
		if err != nil {
			return Entry{}, fmt.Errorf("invalid added date: %w", err)
		}
	}
	if form.Has("read") && form.Get("read") != "" {
		e.Read, err = ParseDate(form.Get("read"))
		if err != nil {
			return Entry{}, fmt.Errorf("invalid read date: %w", err)
		}
	}
	if form.Has("reviewed") && form.Get("reviewed") != "" {
		e.Reviewed, err = ParseDate(form.Get("reviewed"))
		if err != nil {
			return Entry{}, fmt.Errorf("invalid reviewed date: %w", err)
		}
	}

	return e, nil
}
