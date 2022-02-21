package entry

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"runtime"
	"strings"
	"sync"

	sfs "github.com/cceckman/reading-list/server/fs"
	multierror "github.com/hashicorp/go-multierror"
)

// How many cache updates are permitted at a time.
var updateConcurrency int

func init() {
	updateConcurrency = runtime.NumCPU()
}

// Create a new EntryManager for the given directory.
func NewManager(dataDir sfs.CreateFS) (*EntryManager, error) {
	m := &EntryManager{
		dataDir:     dataDir,
		cacheUpdate: make(chan struct{}, 1),
		cacheLock:   sync.Mutex{},
		cache:       make(map[string]*Entry),
	}
	// Initialize the token.
	m.cacheUpdate <- struct{}{}
	if err := m.refreshCache(); err != nil {
		log.Printf("error in doing initial cache refresh: %v", err)
		// However: we continue; data errors shouldn't impact startup.
		// TODO: Better distinction between "data errors" and "real errors"
	}

	return m, nil
}

// EntryManager provides CRUD updates to persisted Entrys.
type EntryManager struct {
	dataDir sfs.CreateFS

	// Token for full-cache updates; we don't want to run >1 at a time.
	cacheUpdate chan struct{}
	// Cache of entries by ID; Mutex-guarded.
	// Entries themselves are considered read-only; any Update should insert a new *Entry
	// rather than modifying an existing one.
	cacheLock sync.Mutex
	cache     map[string]*Entry
}

// Refreshes the (whole) cache.
//
// TODO: Cache invalidation is hard! :-D
// This keeps items that have been deleted.
func (s *EntryManager) refreshCache() error {
	log.Printf("refreshing entry cache from %+v", s.dataDir)
	select {
	case <-s.cacheUpdate:
		log.Printf("refreshing entry cache from %+v", s.dataDir)
		// Token available; proceed with the update.
	default:
		log.Printf("entry cache refresh in progress, waiting...")
		// Token not availble.
		// Wait for it to be available, indicating the in-progress update is done;
		// then return. The other thread has "taken care of" the update we were asked to do.
		s.cacheUpdate <- (<-s.cacheUpdate)
		return nil
	}
	// We have acquired the token. Return it when we're done.
	defer func() {
		s.cacheUpdate <- struct{}{}
	}()

	// Orchestration:
	// - Errors aggregate in an unbuffered channel
	// - Filenames become IDs in one thread
	// - Worker threads multiplex file reads until ids are flushed
	errs := make(chan error)
	ids := make(chan string)
	var wg sync.WaitGroup
	ents := make(chan *Entry, updateConcurrency)
	go func() {
		defer close(ids)
		matches, err := fs.Glob(s.dataDir, "*.md")
		if err != nil {
			errs <- err
			return
		}
		log.Printf("found %d entries in filesystem", len(matches))

		for _, id := range matches {
			ids <- strings.TrimSuffix(id, ".md")
		}
	}()
	wg.Add(updateConcurrency)
	for i := 0; i < updateConcurrency; i++ {
		go func() {
			defer wg.Done()
			for id := range ids {
				if ent, err := s.readItem(id); err != nil {
					errs <- err
				} else {
					ents <- ent
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(ents)
	}()

	newCache := make(map[string]*Entry)
	var result error
Loop:
	for {
		select {
		case err := <-errs:
			result = multierror.Append(result, err)
		case ent, ok := <-ents:
			if !ok {
				break Loop
			}
			newCache[ent.Id] = ent
		}
	}
	// All threads that would write to errs are done.
	close(errs)
	// Flush any remaining errors.
	for err := range errs {
		result = multierror.Append(result, err)
	}
	log.Printf("finished refresh with %d items", len(newCache))
	if result != nil {
		log.Printf("error in processing refresh: %v", result)
		return result
	}

	s.cacheLock.Lock()
	defer s.cacheLock.Unlock()
	s.cache = newCache

	// TODO: This wasn't tested (returning a non-nil error)!
	return result
}

// Read the item with the given ID.
func (s *EntryManager) readItem(id string) (*Entry, error) {
	log.Printf("updating %q", id)
	f, err := s.getFile(id)
	if err != nil {
		return nil, fmt.Errorf("could not read entry: %w", err)
	}
	defer f.Close()
	ent, err := Read(id, f)
	if err != nil {
		return nil, fmt.Errorf("could not parse entry: %w", err)
	}
	if ent.Id != id {
		return nil, fmt.Errorf("failed internal consistency check: Id %q != %q", ent.Id, id)
	}
	return &ent, nil
}

// Read the entry with the given contents out from storage.
// Refreshes the cache entry.
func (s *EntryManager) Read(id string) (*Entry, error) {
	var ent *Entry
	var err error
	if ent, err = s.readItem(id); err != nil {
		return nil, fmt.Errorf("could not read state of %q: %w", id, err)
	}
	s.cacheLock.Lock()
	defer s.cacheLock.Unlock()
	s.cache[ent.Id] = ent
	return ent, nil
}

// Update (or create) the entry.
func (s *EntryManager) Update(e Entry) (*Entry, error) {
	// An invalid ID - such as one with '..' - could throw off our path traversal.
	// Check before doing any creation etc.
	if err := e.ValidID(); err != nil {
		return nil, err
	}

	var f fs.File
	var err error
	if f, err = s.dataDir.Create(e.Id + ".md"); errors.Is(err, fs.ErrExist) {
		// Read the current contents in order to update:
		f, err = s.getFile(e.Id)
		if err != nil {
			return nil, fmt.Errorf("could not read for entry update %q: %w", e.Id, err)
		}
		oldEnt, err := Read(e.Id, f)
		if err != nil {
			return nil, fmt.Errorf("could not parse for update entry %q: %w", e.Id, err)
		}
		e.original = oldEnt.original
		return nil, err
	}
	// Ensure we clean up the file...
	defer func() {
		if f == nil {
			return
		}
		if err := f.Close(); err != nil {
			log.Printf("error closing file for entry %q: %v", e.Id, err)
		}
	}()

	// We've loaded the current contents of the file.
	rwf, ok := f.(sfs.RWFile)
	if !ok {
		return nil, fmt.Errorf("file not available for update for entry %q", e.Id)
	}
	if _, err := rwf.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("could not seek in file for entry %q: %w", e.Id, err)
	}
	if _, err := e.WriteTo(rwf); err != nil {
		return nil, fmt.Errorf("failed to write file for entry %q: %w", e.Id, err)
	}
	// Manually close; flush the writes above.
	err = f.Close()
	f = nil // Prevent duplicate close
	if err != nil {
		return nil, fmt.Errorf("could not close file for entry update: %w", err)
	}

	// Refresh the contents from disk before we consider ourselves complete.
	return s.Read(e.Id)
}

// List up to `limit` entries.
func (s *EntryManager) List(limit int) ([]*Entry, error) {
	// TODO: Actually use the cache as a cache; for now, consume synchronously.
	/*defer func() {
	go func() {*/
	if err := s.refreshCache(); err != nil {
		log.Print("error while doing background cache update: ", err)
		return nil, err
	}
	/*
		}()
	}()*/

	// But don't delay on file IO right now: serve from cache.
	s.cacheLock.Lock()
	ptrs := make([]*Entry, 0, len(s.cache))
	log.Printf("entries: %+v", s.cache)
	for _, ent := range s.cache {
		ptrs = append(ptrs, ent)
	}
	FifoSort(ptrs)
	s.cacheLock.Unlock()
	log.Printf("entries: %+v", ptrs)

	max := len(ptrs)
	if limit < max {
		max = limit
	}
	rs := ptrs[0:max]
	return rs, nil
}

func (s *EntryManager) getFile(id string) (sfs.RWFile, error) {
	name := id + ".md"
	f, err := s.dataDir.Open(name)
	if err != nil {
		return nil, fmt.Errorf("could not open file %s: %w", name, err)
	}
	rwf, ok := f.(sfs.RWFile)
	if !ok {
		return nil, fmt.Errorf("could not treat file as read/write/seek/close (file: %s)", name)
	}
	return rwf, nil
}
