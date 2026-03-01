package audit

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileLogger_JSONL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")

	logger, err := NewFileLogger(path)
	if err != nil {
		t.Fatalf("NewFileLogger: %v", err)
	}

	entry := NewEntry()
	entry.Client = "acme"
	entry.Domain = "vpn"
	entry.ExecPath = "/usr/bin/echo"
	entry.Argv = []string{"hello"}
	entry.ExitCode = 0
	entry.DurationMs = 42

	if err := logger.Log(entry); err != nil {
		t.Fatalf("Log: %v", err)
	}
	if err := logger.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("lines = %d, want 1", len(lines))
	}

	var parsed Entry
	if err := json.Unmarshal([]byte(lines[0]), &parsed); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if parsed.Client != "acme" {
		t.Errorf("client = %q, want acme", parsed.Client)
	}
	if parsed.ExitCode != 0 {
		t.Errorf("exit_code = %d, want 0", parsed.ExitCode)
	}
	if parsed.DurationMs != 42 {
		t.Errorf("duration_ms = %d, want 42", parsed.DurationMs)
	}
}

func TestFileLogger_Append(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")

	// Write first entry
	logger1, _ := NewFileLogger(path)
	e1 := NewEntry()
	e1.Client = "first"
	logger1.Log(e1)
	logger1.Close()

	// Write second entry
	logger2, _ := NewFileLogger(path)
	e2 := NewEntry()
	e2.Client = "second"
	logger2.Log(e2)
	logger2.Close()

	data, _ := os.ReadFile(path)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("lines = %d, want 2", len(lines))
	}
}

func TestStderrLogger_Output(t *testing.T) {
	var buf bytes.Buffer
	logger := &StderrLogger{w: &buf}

	entry := NewEntry()
	entry.Client = "test"
	entry.Domain = "web"

	if err := logger.Log(entry); err != nil {
		t.Fatalf("Log: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"client":"test"`) {
		t.Errorf("output should contain client, got: %s", output)
	}
	if !strings.HasSuffix(output, "\n") {
		t.Error("output should end with newline")
	}
}

func TestNoopLogger(t *testing.T) {
	logger := &NoopLogger{}
	if err := logger.Log(NewEntry()); err != nil {
		t.Errorf("Log: %v", err)
	}
	if err := logger.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

func TestFileLogger_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "dir", "audit.log")

	logger, err := NewFileLogger(path)
	if err != nil {
		t.Fatalf("NewFileLogger: %v", err)
	}
	defer logger.Close()

	entry := NewEntry()
	entry.Client = "test"
	if err := logger.Log(entry); err != nil {
		t.Fatalf("Log: %v", err)
	}
}
