package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadIncludeNodes_SingleFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "extra.yaml")
	os.WriteFile(path, []byte(`clients:
  extra:
    domains:
      test:
        dir: "/tmp"
`), 0644)

	nodes, err := loadIncludeNodes([]string{path})
	if err != nil {
		t.Fatalf("loadIncludeNodes: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("nodes = %d, want 1", len(nodes))
	}
}

func TestLoadIncludeNodes_Glob(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.yaml", "b.yaml", "c.txt"} {
		os.WriteFile(filepath.Join(dir, name), []byte("version: 1\n"), 0644)
	}

	pattern := filepath.Join(dir, "*.yaml")
	nodes, err := loadIncludeNodes([]string{pattern})
	if err != nil {
		t.Fatalf("loadIncludeNodes: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("nodes = %d, want 2", len(nodes))
	}
}

func TestLoad_WithIncludes(t *testing.T) {
	dir := t.TempDir()

	// Write include file
	includeDir := filepath.Join(dir, "includes")
	os.MkdirAll(includeDir, 0755)
	os.WriteFile(filepath.Join(includeDir, "extra.yaml"), []byte(`
clients:
  extra:
    domains:
      web:
        dir: "/tmp/extra"
`), 0644)

	// Write main config
	mainPath := filepath.Join(dir, "config.yaml")
	os.WriteFile(mainPath, []byte(`
version: 1
includes:
  - "`+includeDir+`/*.yaml"
clients:
  acme:
    domains:
      vpn:
        dir: "/tmp/acme"
`), 0644)

	cfg, err := Load(mainPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if _, ok := cfg.Clients["acme"]; !ok {
		t.Error("missing client 'acme'")
	}
	if _, ok := cfg.Clients["extra"]; !ok {
		t.Error("missing client 'extra' from include")
	}
}
