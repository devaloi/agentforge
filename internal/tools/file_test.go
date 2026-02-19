package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestReadFileExecute(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("hello world"), 0o644)

	rf := NewReadFile(dir)
	out, err := rf.Execute(context.Background(), map[string]any{"path": "test.txt"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if out != "hello world" {
		t.Errorf("output = %q, want %q", out, "hello world")
	}
}

func TestReadFileSandbox(t *testing.T) {
	rf := NewReadFile("/tmp/sandbox")
	_, err := rf.Execute(context.Background(), map[string]any{"path": "../../../etc/passwd"})
	if err == nil {
		t.Fatal("expected sandbox escape error")
	}
}

func TestReadFileAbsolutePath(t *testing.T) {
	rf := NewReadFile("/tmp/sandbox")
	_, err := rf.Execute(context.Background(), map[string]any{"path": "/etc/passwd"})
	if err == nil {
		t.Fatal("expected error for absolute path")
	}
}

func TestReadFileMissingPath(t *testing.T) {
	rf := NewReadFile("/tmp")
	_, err := rf.Execute(context.Background(), map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing path")
	}
}

func TestWriteFileExecute(t *testing.T) {
	dir := t.TempDir()
	wf := NewWriteFile(dir)

	_, err := wf.Execute(context.Background(), map[string]any{
		"path":    "output.txt",
		"content": "generated content",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "output.txt"))
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(data) != "generated content" {
		t.Errorf("content = %q, want %q", string(data), "generated content")
	}
}

func TestWriteFileCreateSubdir(t *testing.T) {
	dir := t.TempDir()
	wf := NewWriteFile(dir)

	_, err := wf.Execute(context.Background(), map[string]any{
		"path":    "sub/dir/file.go",
		"content": "package main",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	_, err = os.Stat(filepath.Join(dir, "sub/dir/file.go"))
	if err != nil {
		t.Errorf("file not created: %v", err)
	}
}

func TestWriteFileSandbox(t *testing.T) {
	wf := NewWriteFile("/tmp/sandbox")
	_, err := wf.Execute(context.Background(), map[string]any{
		"path":    "../../etc/evil",
		"content": "hacked",
	})
	if err == nil {
		t.Fatal("expected sandbox escape error")
	}
}

func TestWriteFileMissingParams(t *testing.T) {
	wf := NewWriteFile("/tmp")
	_, err := wf.Execute(context.Background(), map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing params")
	}
}
