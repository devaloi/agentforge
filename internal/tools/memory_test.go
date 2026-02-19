package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/devaloi/agentforge/internal/memory"
)

func TestMemoryWriteAndRead(t *testing.T) {
	store := memory.NewStore()
	mw := NewMemoryWrite(store, "researcher")
	mr := NewMemoryRead(store)

	out, err := mw.Execute(context.Background(), map[string]any{
		"key":   "findings",
		"value": "Go is great for APIs",
	})
	if err != nil {
		t.Fatalf("MemoryWrite: %v", err)
	}

	var writeResult map[string]any
	json.Unmarshal([]byte(out), &writeResult)
	if writeResult["success"] != true {
		t.Error("write should succeed")
	}

	out, err = mr.Execute(context.Background(), map[string]any{"key": "findings"})
	if err != nil {
		t.Fatalf("MemoryRead: %v", err)
	}

	var readResult map[string]any
	json.Unmarshal([]byte(out), &readResult)
	if readResult["found"] != true {
		t.Error("key should be found")
	}
	if readResult["value"] != "Go is great for APIs" {
		t.Errorf("value = %q", readResult["value"])
	}
}

func TestMemoryReadNotFound(t *testing.T) {
	store := memory.NewStore()
	mr := NewMemoryRead(store)

	out, err := mr.Execute(context.Background(), map[string]any{"key": "missing"})
	if err != nil {
		t.Fatalf("MemoryRead: %v", err)
	}

	var result map[string]any
	json.Unmarshal([]byte(out), &result)
	if result["found"] != false {
		t.Error("key should not be found")
	}
}

func TestMemoryReadMissingKey(t *testing.T) {
	store := memory.NewStore()
	mr := NewMemoryRead(store)
	_, err := mr.Execute(context.Background(), map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestMemoryWriteMissingParams(t *testing.T) {
	store := memory.NewStore()
	mw := NewMemoryWrite(store, "agent")
	_, err := mw.Execute(context.Background(), map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing params")
	}
}

func TestMemoryWriteAttribution(t *testing.T) {
	store := memory.NewStore()
	mw := NewMemoryWrite(store, "coder")

	mw.Execute(context.Background(), map[string]any{
		"key":   "code",
		"value": "package main",
	})

	entry, ok := store.GetEntry("code")
	if !ok {
		t.Fatal("entry not found")
	}
	if entry.Author != "coder" {
		t.Errorf("Author = %q, want %q", entry.Author, "coder")
	}
}
