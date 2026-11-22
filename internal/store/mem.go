package store

import (
	"context"
	"sync"
	"time"
)

// Memory is an in-process Store for local dev and tests. It forgets
// everything on restart and is refused in production.
type Memory struct {
	mu      sync.Mutex
	visits  map[string]bool
	total   int64
	entries []memEntry
	nextID  int64
	now     func() time.Time
}

type memEntry struct {
	Entry
	visitor string
	deleted bool
}

// NewMemory returns an empty Memory store.
func NewMemory() *Memory {
	return &Memory{visits: map[string]bool{}, now: time.Now}
}

func (m *Memory) Visit(_ context.Context, day time.Time, visitor string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := Day(day).Format("2006-01-02") + "|" + visitor
	if !m.visits[k] {
		m.visits[k] = true
		m.total++
	}
	return m.total, nil
}

func (m *Memory) Visitors(context.Context) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.total, nil
}

func (m *Memory) AddEntry(_ context.Context, e Entry, visitor string) (Entry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextID++
	e.ID = m.nextID
	e.Created = m.now().UTC()
	m.entries = append(m.entries, memEntry{Entry: e, visitor: visitor})
	return e, nil
}

func (m *Memory) Entries(_ context.Context, limit int) ([]Entry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []Entry
	for i := len(m.entries) - 1; i >= 0 && len(out) < limit; i-- {
		if !m.entries[i].deleted {
			out = append(out, m.entries[i].Entry)
		}
	}
	return out, nil
}

func (m *Memory) SignedSince(_ context.Context, visitor string, t time.Time) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	for _, e := range m.entries {
		if e.visitor == visitor && !e.Created.Before(t) {
			n++
		}
	}
	return n, nil
}

func (m *Memory) RecentDuplicate(_ context.Context, message string, t time.Time) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, e := range m.entries {
		if e.Message == message && !e.Created.Before(t) {
			return true, nil
		}
	}
	return false, nil
}

func (m *Memory) DeleteEntry(_ context.Context, id int64) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.entries {
		if m.entries[i].ID == id && !m.entries[i].deleted {
			m.entries[i].deleted = true
			return true, nil
		}
	}
	return false, nil
}

func (m *Memory) Close() {}
