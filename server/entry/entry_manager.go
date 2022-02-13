package entry

import (
	"fmt"
	"io"
	"io/fs"
)

// Create a new EntryManager for the given directory.
func NewManager(dataDir fs.FS) *EntryManager {
	return &EntryManager{
		dataDir: dataDir,
	}
}

// An object like a read-write file.
type rwfile interface {
	fs.File
	io.Reader
	io.Writer
	io.Closer
	io.Seeker
}

// EntryManager provides CRUD updates to persisted Entrys.
type EntryManager struct {
	dataDir fs.FS
}

// Read the entry with the given contents out from storage.
func (s *EntryManager) Read(id string) (*Entry, error) {
	f, err := s.getFile(id)
	if err != nil {
		return nil, fmt.Errorf("could not read entry: %w", err)
	}
	ent, err := Read(id, f)
	if err != nil {
		return nil, fmt.Errorf("could not parse entry: %w", err)
	}
	return ent, nil
}

// Update (or create) the entry.
func (s *EntryManager) Update(*Entry) error {
	return fmt.Errorf("unimplemented")
}

func (s *EntryManager) List(limit int) ([]*Entry, error) {
	matches, err := fs.Glob(s.dataDir, "*.md")
	if err != nil {
		return nil, fmt.Errorf("could not list entries: %w", err)
	}
	_ = matches
	return nil, fmt.Errorf("unimplemented")
}

func (s *EntryManager) getFile(id string) (rwfile, error) {
	name := id + ".md"
	f, err := s.dataDir.Open(name)
	if err != nil {
		return nil, fmt.Errorf("could not open file %s: %w", name, err)
	}
	rwf, ok := f.(rwfile)
	if !ok {
		return nil, fmt.Errorf("could not treat file as read/write/seek/close (file: %s)", name)
	}
	return rwf, nil
}
