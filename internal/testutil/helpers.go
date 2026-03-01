package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// WriteConfig writes a YAML config to a temp file and returns the path.
func WriteConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

// WriteExecutable creates a temp executable script and returns the directory and filename.
func WriteExecutable(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}
	return dir
}

// WriteExecutableIn creates an executable in the specified directory.
func WriteExecutableIn(t *testing.T, dir, name, content string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}
	return path
}
