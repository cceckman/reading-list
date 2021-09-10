package entry

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"
	"unicode"

	"github.com/spf13/afero"
)

// Create a new EntryManager for the given directory.
func NewManager(dataDir afero.Fs) EntryManager {
	return EntryManager{
		dataDir: &afero.Afero{Fs: dataDir},
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
	dataDir *afero.Afero
}

// Write an entry out to storage.
// We require a separate "key" so that the entry's title can change without creating a duplicate entry.
func (s *EntryManager) write(w io.WriteCloser, fm *FrontMatter, body io.Reader) error {
	if _, err := fm.WriteTo(w); err != nil {
		return fmt.Errorf("could not write front matter: %w", err)
	}
	if body != nil {
		if _, err := io.Copy(w, body); err != nil {
			return fmt.Errorf("could not write body: %w", err)
		}
	}

	return w.Close()
}

func (s *EntryManager) read(key string) (*FrontMatter, *bytes.Buffer, error) {
	b, err := afero.ReadFile(s.dataDir, filename(key))
	if err != nil {
		return nil, nil, fmt.Errorf("could not read file: %w", err)
	}
	entry, body, err := Read(bytes.NewBuffer(b))
	if err != nil {
		return nil, nil, fmt.Errorf("could not parse filecontents: %w", err)
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
		return "", fmt.Errorf("could not parse shared URL: %w", err)
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
	fn := filename(key)

	if exists, err := s.dataDir.Exists(fn); err != nil {
		return "", err
	} else if exists {
		return "", afero.ErrFileExists
	}

	f, err := s.dataDir.Create(fn)
	if err != nil {
		return key, fmt.Errorf("could not create file: %w", err)
	}
	return key, s.write(f, fm, nil)
}

func (s *EntryManager) Read(key string) (*FrontMatter, error) {
	entry, _, err := s.read(key)
	return entry, err
}
