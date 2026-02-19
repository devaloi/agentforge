package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/devaloi/agentforge/internal/provider"
)

// WriteFile writes content to a file within a sandboxed directory.
type WriteFile struct {
	baseDir string
}

// NewWriteFile creates a WriteFile tool sandboxed to the given directory.
func NewWriteFile(baseDir string) *WriteFile {
	return &WriteFile{baseDir: baseDir}
}

func (w *WriteFile) Name() string        { return "write_file" }
func (w *WriteFile) Description() string { return "Write content to a file in the output directory" }

func (w *WriteFile) Schema() provider.JSONSchema {
	return NewSchemaBuilder().
		AddString("path", "Relative file path to write", true).
		AddString("content", "File content to write", true).
		Build()
}

func (w *WriteFile) Execute(_ context.Context, params map[string]any) (string, error) {
	path, _ := params["path"].(string)
	content, _ := params["content"].(string)

	if path == "" || content == "" {
		return "", fmt.Errorf("write_file: 'path' and 'content' are required")
	}

	resolved, err := w.resolve(path)
	if err != nil {
		return "", err
	}

	dir := filepath.Dir(resolved)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("write_file: create directory: %w", err)
	}

	if err := os.WriteFile(resolved, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("write_file: %w", err)
	}

	result := map[string]any{"success": true, "path": path}
	data, _ := json.Marshal(result)
	return string(data), nil
}

func (w *WriteFile) resolve(path string) (string, error) {
	clean := filepath.Clean(path)
	if filepath.IsAbs(clean) || strings.Contains(clean, "..") {
		return "", fmt.Errorf("write_file: path %q escapes sandbox", path)
	}
	return filepath.Join(w.baseDir, clean), nil
}
