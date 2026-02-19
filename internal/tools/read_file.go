package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/devaloi/agentforge/internal/provider"
)

// ReadFile reads a file from within a sandboxed directory.
type ReadFile struct {
	baseDir string
}

// NewReadFile creates a ReadFile tool sandboxed to the given directory.
func NewReadFile(baseDir string) *ReadFile {
	return &ReadFile{baseDir: baseDir}
}

func (r *ReadFile) Name() string        { return "read_file" }
func (r *ReadFile) Description() string { return "Read file contents from the output directory" }

func (r *ReadFile) Schema() provider.JSONSchema {
	return NewSchemaBuilder().
		AddString("path", "Relative file path to read", true).
		Build()
}

func (r *ReadFile) Execute(_ context.Context, params map[string]any) (string, error) {
	path, _ := params["path"].(string)
	if path == "" {
		return "", fmt.Errorf("read_file: 'path' is required")
	}

	resolved, err := r.resolve(path)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(resolved)
	if err != nil {
		return "", fmt.Errorf("read_file: %w", err)
	}

	return string(data), nil
}

func (r *ReadFile) resolve(path string) (string, error) {
	clean := filepath.Clean(path)
	if filepath.IsAbs(clean) || strings.Contains(clean, "..") {
		return "", fmt.Errorf("read_file: path %q escapes sandbox", path)
	}
	return filepath.Join(r.baseDir, clean), nil
}
