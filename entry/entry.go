package entry

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"time"

	yaml "gopkg.in/yaml.v3"
)

// Metadata for a reading-list entry.
type Entry struct {
	// Web source: how can this item be found on the internet?
	//
	// Note that this is optional to allow for e.g. "book source",
	// where the canonical reference may be an ISBN.
	WebSource *Source `yaml:"web-source,omitempty"`

	// Who wrote this work, and where can they be found?
	Author *Source `yaml:",omitempty"`

	// When this entry was moved from "in the queue" to "read".
	Read time.Time `yaml:",omitempty"`
	// When commentary on this entry was made available.
	Reviewed time.Time `yaml:",omitempty"`

	// Discovery data: how did I come across this item?
	// This may be rendered as "found via..."
	Discovery *Source `yaml:",omitempty"`
}

// Discovery / link metadata.
type Source struct {
	Text   string
	RawUrl string `yaml:"url,omitempty"`
}

func (s *Source) Url() (*url.URL, error) {
	return url.Parse(s.RawUrl)
}

// Hugo front matter, with a reading-list subobject.
type FrontMatter struct {
	// Title for the underlying work; also used as the page title.
	Title string

	// Summary of the underlying work.
	// We use Hugo's metadata field for this rather than the "<!-- more -->" delimiter
	// to make the content vs. summary division cleaner.
	Summary string `yaml:",omitempty"`

	// Time (date) at which this entry was enqueued in the reading list.
	Date time.Time `yaml:",omitempty"`

	// Whether this item should remain a draft, i.e. not rendered.
	Draft bool

	// ReadingList metadata.
	ReadingList *Entry `yaml:"reading-list,omitempty"`

	// Any other key-value entries included in the front matter.
	// `,inline` allows us to preserve them on encode / decode, without actually caring
	// about their contents.
	Other map[string]interface{} `yaml:",inline"`
}

const (
	// Header: the first bytes of a document with (YAML) front matter.
	header = "---\n"
	// Delimiter: the delimiter between (YAML) front matter and the main document.
	delimiter = "\n---\n"
)

// Write the front-matter to the output channel.
func (fm *FrontMatter) WriteTo(w io.Writer) (n int64, err error) {
	// To return an accurate value for (n), write to an in-memory buffer first.
	// Encoder does not write leading/trailing document markers unless multiple documents
	// are written; begin by writing a start-of-doc marker.
	buf := bytes.NewBufferString(header)
	enc := yaml.NewEncoder(buf)
	enc.SetIndent(2)
	if err = enc.Encode(fm); err != nil {
		return
	}
	if _, err = buf.WriteString(delimiter); err != nil {
		return
	}
	return buf.WriteTo(w)
}

// Read the front-matter and body from the input channel.
func Read(r io.Reader) (*FrontMatter, []byte, error) {
	rd := bytes.NewBuffer(nil)
	if _, err := rd.ReadFrom(r); err != nil {
		return nil, nil, err
	}
	buf := rd.Bytes()
	if !bytes.HasPrefix(buf, []byte(header)) {
		return nil, nil, fmt.Errorf("did not find front-matter prefix")
	}
	bytes.TrimPrefix(buf, []byte(header))

	/*
		It turns out we can just `bytes.Split` to get before/after the front-matter separator-
		even if the document's contents include the delimiter string "\n---\n".
		In YAML, if some field `example` has the content "\n---\n", it must be encoded as a multiline string.
		Per https://yaml-multiline.info, each successive line of a multiline string must be indented
		by some nonzero amount.
		If I understand correctly, this means the sequence "\n--" is syntactically invalid:
		it's not a list-item (only one "-"), and within a multiline string, there would need to be at least one
		space between "\n" and the first "-".

		The sequence "\n---\n" might appear in the body of the document - but we can limit ourselves to only
		finding the first occurrence, and treating everything after it as the main contents.
	*/

	var sections [][]byte
	if sections = bytes.SplitN(buf, []byte(delimiter), 2); len(sections) != 2 {
		return nil, nil, fmt.Errorf("did not find front-matter end delimiter")
	}
	front := sections[0]
	body := sections[1]
	f := &FrontMatter{}
	if err := yaml.Unmarshal(front, f); err != nil {
		return nil, nil, err
	}
	return f, body, nil
}
