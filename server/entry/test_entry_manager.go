package entry

import (
	"fmt"
)

// TestEntryManager provides a fake of the server.EntryManager type for testing.
type TestEntryManager struct {
	Items     map[string]*Entry
	ListError error
}

func (em *TestEntryManager) List(limit int) ([]*Entry, error) {
	if em.ListError != nil {
		return nil, em.ListError
	}
	if limit > len(em.Items) {
		limit = len(em.Items)
	}
	items := make([]*Entry, 0, limit)
	for _, v := range em.Items {
		items = append(items, v)
		if len(items) >= limit {
			break
		}
	}
	return items, nil
}
func (em *TestEntryManager) Read(id string) (*Entry, error) {
	if e, ok := em.Items[id]; !ok {
		return nil, fmt.Errorf("not found")
	} else {
		return e, nil
	}
}

func (em *TestEntryManager) Update(e *Entry) error {
	if e.Id == "" {
		return fmt.Errorf("no ID provided")
	}
	em.Items[e.Id] = e
	return nil
}
