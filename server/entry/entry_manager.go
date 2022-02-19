package entry

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"runtime"
	"strings"
	"sync"

	multierror "github.com/hashicorp/go-multierror"
)

// How many cache updates are permitted at a time.
var updateConcurrency int

func init() {
	updateConcurrency = runtime.NumCPU()
}

// Create a new EntryManager for the given directory.
func NewManager(dataDir fs.FS) (*EntryManager, error) {
	m := &EntryManager{
		dataDir:     dataDir,
		cacheUpdate: make(chan struct{}, 1),
		cacheLock:   sync.Mutex{},
		cache:       make(map[string]*Entry),
	}
	// Initialize the token.
	m.cacheUpdate <- struct{}{}
	if err := m.refreshCache(); err != nil {
		return nil, err
	}

	return m, nil
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

	// Token for full-cache updates; we don't want to run >1 at a time.
	cacheUpdate chan struct{}
	// Cache of entries by ID; Mutex-guarded.
	// Entries themselves are considered read-only; any Update should insert a new *Entry
	// rather than modifying an existing one.
	cacheLock sync.Mutex
	cache     map[string]*Entry
}

// Refreshes the (whole) cache.
func (s *EntryManager) refreshCache() error {
	select {
	case <-s.cacheUpdate:
		// Token available; proceed with the update.
	default:
		// Token not availble.
		// Wait for it to be available, indicating the in-progress update is done;
		// then return. The other thread has "taken care of" the update we were asked to do.
		s.cacheUpdate <- (<-s.cacheUpdate)
	}
	// We have acquired the token. Return it when we're done.
	defer func() {
		s.cacheUpdate <- struct{}{}
	}()

	errs := make(chan error)
	ids := make(chan string)

	// One thread for the "list" operation.
	go func() {
		defer close(ids)
		matches, err := fs.Glob(s.dataDir, "*.md")
		if err != nil {
			errs <- err
			return
		}

		for _, id := range matches {
			ids <- strings.TrimSuffix(id, ".md")
		}
	}()

	// Update cache with several threads.
	// These will wind up doing file I/O; limit the total number of threads.
	var wg sync.WaitGroup
	for i := 0; i < updateConcurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range ids {
				if err := s.refreshCacheItem(id); err != nil {
					errs <- err
				}
			}
		}()
	}
	// Close error channel when all update threads are done.
	// The "list" thread is guaranteed to send to errs before closing ids,
	// so any errors from that thread are necessarily captured.
	go func() {
		wg.Wait()
		close(errs)
	}()

	var result error
	for err := range errs {
		result = multierror.Append(result, err)
	}

	return nil
}

// Refresh the cache entry for the given ID.
func (s *EntryManager) refreshCacheItem(id string) error {
	f, err := s.getFile(id)
	if err != nil {
		return fmt.Errorf("could not read entry: %w", err)
	}
	ent, err := Read(id, f)
	if err != nil {
		return fmt.Errorf("could not parse entry: %w", err)
	}

	s.cacheLock.Lock()
	defer s.cacheLock.Unlock()
	s.cache[ent.Id] = &ent
	return nil
}

// Read the entry with the given contents out from storage.
func (s *EntryManager) Read(id string) (*Entry, error) {
	if err := s.refreshCacheItem(id); err != nil {
		return nil, fmt.Errorf("could not read state of %q: %w", id, err)
	}
	s.cacheLock.Lock()
	defer s.cacheLock.Unlock()
	if ent, ok := s.cache[id]; ok {
		return ent, nil
	} else {
		return nil, fmt.Errorf("no entry found with id %q", id)
	}
}

// Update (or create) the entry.
func (s *EntryManager) Update(*Entry) error {
	return fmt.Errorf("unimplemented")
}

// List up to `limit` entries.
//
// This implementation serves from the local cache, but
func (s *EntryManager) List(limit int) ([]*Entry, error) {
	// When we're done, refresh the cache in a background thread.
	defer func() {
		go func() {
			if err := s.refreshCache(); err != nil {
				log.Print("error while doing background cache update: ", err)
			}
		}()
	}()

	// But don't delay on file IO right now: serve from cache.
	s.cacheLock.Lock()
	ptrs := make([]*Entry, 0, len(s.cache))
	for _, ent := range s.cache {
		ptrs = append(ptrs, ent)
	}
	FifoSort(ptrs)
	s.cacheLock.Unlock()

	max := len(ptrs)
	if limit < max {
		max = limit
	}
	return ptrs[0:max], nil
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
