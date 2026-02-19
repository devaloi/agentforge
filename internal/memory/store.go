package memory

import (
	"sync"
	"time"
)

// Store is a thread-safe key-value store for agent collaboration.
type Store struct {
	mu      sync.RWMutex
	entries map[string]Entry
}

// NewStore creates an empty shared memory store.
func NewStore() *Store {
	return &Store{entries: make(map[string]Entry)}
}

// Read retrieves a value by key. Returns the value and whether it was found.
func (s *Store) Read(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.entries[key]
	if !ok {
		return "", false
	}
	return entry.Value, true
}

// Write stores a value with agent attribution.
func (s *Store) Write(key, value, author string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[key] = Entry{
		Key:       key,
		Value:     value,
		Author:    author,
		Timestamp: time.Now(),
	}
}

// Delete removes a key from the store.
func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, key)
}

// List returns all entries in the store.
func (s *Store) List() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entries := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		entries = append(entries, e)
	}
	return entries
}

// Keys returns all keys in the store.
func (s *Store) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.entries))
	for k := range s.entries {
		keys = append(keys, k)
	}
	return keys
}

// GetEntry retrieves the full entry including metadata.
func (s *Store) GetEntry(key string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.entries[key]
	return entry, ok
}
