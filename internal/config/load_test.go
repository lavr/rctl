package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
version: 1
defaults:
  env:
    LANG: "C"
  commands_path:
    - "~/.local/share/rctl/commands"
clients:
  acme:
    domains:
      vpn:
        dir: "~/work/acme/vpn"
        env:
          VPN_ENV: "prod"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Version != 1 {
		t.Errorf("version = %d, want 1", cfg.Version)
	}
	if len(cfg.Clients) != 1 {
		t.Errorf("clients count = %d, want 1", len(cfg.Clients))
	}
	client := cfg.Clients["acme"]
	if client == nil {
		t.Fatal("client 'acme' not found")
	}
	domain := client.Domains["vpn"]
	if domain.Env["VPN_ENV"] != "prod" {
		t.Errorf("domain env VPN_ENV = %q, want %q", domain.Env["VPN_ENV"], "prod")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoad_InvalidVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `version: 99`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("expected error for invalid version")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `{{{invalid`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoad_TildeExpansion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
version: 1
defaults:
  commands_path:
    - "~/commands"
clients:
  foo:
    domains:
      bar:
        dir: "~/work/foo"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	home, _ := os.UserHomeDir()

	// Check defaults commands_path expanded
	if len(cfg.Defaults.CommandsPath) != 1 {
		t.Fatalf("commands_path len = %d, want 1", len(cfg.Defaults.CommandsPath))
	}
	expected := filepath.Join(home, "commands")
	if cfg.Defaults.CommandsPath[0] != expected {
		t.Errorf("commands_path[0] = %q, want %q", cfg.Defaults.CommandsPath[0], expected)
	}

	// Check domain dir expanded
	domain := cfg.Clients["foo"].Domains["bar"]
	expectedDir := filepath.Join(home, "work/foo")
	if domain.Dir != expectedDir {
		t.Errorf("domain dir = %q, want %q", domain.Dir, expectedDir)
	}
}
