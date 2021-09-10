package entry

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"strings"
	"time"
	"unicode"
)

// Create a new EntryManager for the given directory.
func NewManager(dataDir fs.FS) EntryManager {
	return EntryManager{
		dataDir: dataDir,
	}
}

func slug(title string) string {
	lower := strings.ToLower(title)
	return strings.Map(func(c rune) rune {
		if unicode.IsLetter(c) {
			return c
		} else {
			return '-'
		}
	}, lower)
}

func filename(slug string) string {
	return slug + ".md"
}

// EntryManager provides CRUD updates to persisted Entrys.
type EntryManager struct {
	dataDir fs.FS
}

// Write an entry out to storage.
// We require a separate "key" so that the entry's title can change without creating a duplicate entry.
func (s *EntryManager) write(key string, fm *FrontMatter, body io.Reader) error {
	fn := filename(key)
	f, err := s.dataDir.Open(fn)
	if err != nil {
		return fmt.Errorf("could not open file %s to write: %e", fn, err)
	}

	// We don't know how "writeable" the thing that came from fs.FS is,
	// so we have to duck-type our way through writing.
	writeable, ok := f.(io.Writer)
	if !ok {
		return fmt.Errorf("could not get file %s as writeable: got: %T, want: io.Writer", fn, f)
	}
	if _, err := fm.WriteTo(writeable); err != nil {
		return fmt.Errorf("could not write front matter: %e", err)
	}
	if _, err := io.Copy(writeable, body); err != nil {
		return fmt.Errorf("could not write body: %e", err)
	}

	if closer, ok := f.(io.Closer); ok {
		return closer.Close()
	} else {
		return nil
	}
}

func (s *EntryManager) read(key string) (*FrontMatter, *bytes.Buffer, error) {
	b, err := fs.ReadFile(s.dataDir, filename(key))
	if err != nil {
		return nil, nil, fmt.Errorf("could not read file: %e", err)
	}
	entry, body, err := Read(bytes.NewBuffer(b))
	if err != nil {
		return nil, nil, fmt.Errorf("could not parse filecontents: %e", err)
	}
	return entry, bytes.NewBuffer(body), nil
}

// Create a stub entry as described by the arguments
func (s *EntryManager) Create(title, rawUrl string) (string, error) {
	if title == "" || rawUrl == "" {
		return "", fmt.Errorf("missing expected parameter: title: %q rawUrl: %q", title, rawUrl)
	}

	u, err := url.Parse(rawUrl)
	if err != nil {
		return "", fmt.Errorf("could not parse shared URL: %e", err)
	}
	host := u.Hostname()

	fm := &FrontMatter{
		Title: title,
		Date:  time.Now(),
		Draft: true,
		ReadingList: &Entry{
			WebSource: &Source{
				Text:   fmt.Sprintf("at %s", host),
				RawUrl: u.String(),
			},
		},
	}
	// The key is the initial title of the entry, slug-ified
	key := slug(fm.Title)
	// Create a front-matter-only entry; no body
	return key, s.write(key, fm, bytes.NewBufferString(""))
}

func (s *EntryManager) Read(key string) (*FrontMatter, error) {
	entry, _, err := s.read(key)
	return entry, err
}

func (s *EntryManager) Update(key string, fm *FrontMatter) error {
	_, body, err := s.read(key)
	if err != nil {
		return fmt.Errorf("could not read entry for update: %e", err)
	}
	if err := s.write(key, fm, body); err != nil {
		return fmt.Errorf("could not write entry for update: %e", err)
	}
	return nil
}
