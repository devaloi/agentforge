package memory

import (
	"sync"
	"testing"
)

func TestStoreReadWrite(t *testing.T) {
	s := NewStore()
	s.Write("key1", "value1", "researcher")

	val, ok := s.Read("key1")
	if !ok {
		t.Fatal("key not found")
	}
	if val != "value1" {
		t.Errorf("value = %q, want %q", val, "value1")
	}
}

func TestStoreReadNotFound(t *testing.T) {
	s := NewStore()
	_, ok := s.Read("nonexistent")
	if ok {
		t.Error("should not find nonexistent key")
	}
}

func TestStoreDelete(t *testing.T) {
	s := NewStore()
	s.Write("key1", "value1", "agent")
	s.Delete("key1")

	_, ok := s.Read("key1")
	if ok {
		t.Error("key should be deleted")
	}
}

func TestStoreList(t *testing.T) {
	s := NewStore()
	s.Write("a", "1", "agent1")
	s.Write("b", "2", "agent2")

	entries := s.List()
	if len(entries) != 2 {
		t.Errorf("List len = %d, want 2", len(entries))
	}
}

func TestStoreKeys(t *testing.T) {
	s := NewStore()
	s.Write("x", "1", "agent")
	s.Write("y", "2", "agent")

	keys := s.Keys()
	if len(keys) != 2 {
		t.Errorf("Keys len = %d, want 2", len(keys))
	}
}

func TestStoreAttribution(t *testing.T) {
	s := NewStore()
	s.Write("findings", "data", "researcher")

	entry, ok := s.GetEntry("findings")
	if !ok {
		t.Fatal("entry not found")
	}
	if entry.Author != "researcher" {
		t.Errorf("Author = %q, want %q", entry.Author, "researcher")
	}
	if entry.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

func TestStoreConcurrent(t *testing.T) {
	s := NewStore()
	var wg sync.WaitGroup

	for i := range 100 {
		wg.Add(2)
		go func(n int) {
			defer wg.Done()
			s.Write("key", "value", "agent")
		}(i)
		go func(n int) {
			defer wg.Done()
			s.Read("key")
		}(i)
	}

	wg.Wait()
}

func TestStoreOverwrite(t *testing.T) {
	s := NewStore()
	s.Write("key", "v1", "agent1")
	s.Write("key", "v2", "agent2")

	val, ok := s.Read("key")
	if !ok {
		t.Fatal("key not found")
	}
	if val != "v2" {
		t.Errorf("value = %q, want %q", val, "v2")
	}

	entry, _ := s.GetEntry("key")
	if entry.Author != "agent2" {
		t.Errorf("Author = %q, want %q", entry.Author, "agent2")
	}
}
