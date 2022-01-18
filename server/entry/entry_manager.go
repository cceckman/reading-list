package entry

import (
	"fmt"
	"io"
	"io/fs"
	"strings"
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
func (s *EntryManager) write(w io.WriteCloser, e *Entry, body io.Reader) error {
	return fmt.Errorf("unimplemented")
}

// Read the entry with the given contents out from storage.
func (s *EntryManager) Read(id string) (*Entry, error) {
	return nil, fmt.Errorf("unimplemented")
}

// Create a stub entry as described by the arguments.
// Returns the ID of the new item.
func (s *EntryManager) Create(title, rawUrl string) (string, error) {
	return "", fmt.Errorf("unimplemented")
}
