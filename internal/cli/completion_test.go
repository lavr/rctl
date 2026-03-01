package cli

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/spf13/cobra"
)

func TestCompleteClients(t *testing.T) {
	cfgPath, _ := writeTestConfig(t)
	app, _, _ := newTestApp(t, cfgPath)

	results, directive := completeClients(app, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("directive = %v, want NoFileComp", directive)
	}

	// Should contain "acme" and alias "ac"
	if !contains(results, "acme") {
		t.Errorf("results = %v, should contain 'acme'", results)
	}
	if !contains(results, "ac") {
		t.Errorf("results = %v, should contain alias 'ac'", results)
	}
}

func TestCompleteClients_NilConfig(t *testing.T) {
	app := NewApp("test")
	results, directive := completeClients(app, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("directive = %v, want NoFileComp", directive)
	}
	if len(results) != 0 {
		t.Errorf("results = %v, want empty", results)
	}
}

func TestCompleteDomains(t *testing.T) {
	cfgPath, _ := writeTestConfig(t)
	app, _, _ := newTestApp(t, cfgPath)

	results, directive := completeDomains(app, "acme", "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("directive = %v, want NoFileComp", directive)
	}

	// Should contain "vpn", "vpn-staging", "ansible", and alias "v"
	if !contains(results, "vpn") {
		t.Errorf("results = %v, should contain 'vpn'", results)
	}
	if !contains(results, "vpn-staging") {
		t.Errorf("results = %v, should contain 'vpn-staging'", results)
	}
	if !contains(results, "ansible") {
		t.Errorf("results = %v, should contain 'ansible'", results)
	}
	if !contains(results, "v") {
		t.Errorf("results = %v, should contain alias 'v'", results)
	}
}

func TestCompleteDomains_WithAlias(t *testing.T) {
	cfgPath, _ := writeTestConfig(t)
	app, _, _ := newTestApp(t, cfgPath)

	// Use client alias "ac"
	results, _ := completeDomains(app, "ac", "")
	if !contains(results, "vpn") {
		t.Errorf("results = %v, should contain 'vpn' (via alias)", results)
	}
}

func TestCompleteDomains_UnknownClient(t *testing.T) {
	cfgPath, _ := writeTestConfig(t)
	app, _, _ := newTestApp(t, cfgPath)

	results, directive := completeDomains(app, "nonexistent", "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("directive = %v, want NoFileComp", directive)
	}
	if len(results) != 0 {
		t.Errorf("results = %v, want empty", results)
	}
}

func TestCompleteCommands(t *testing.T) {
	cfgPath, _ := writeTestConfig(t)
	app, _, _ := newTestApp(t, cfgPath)

	results, directive := completeCommands(app, "acme", "vpn", "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("directive = %v, want NoFileComp", directive)
	}

	// Should contain profile "up"
	if !contains(results, "up") {
		t.Errorf("results = %v, should contain profile 'up'", results)
	}

	// Should contain builtin "kctx"
	if !contains(results, "kctx") {
		t.Errorf("results = %v, should contain builtin 'kctx'", results)
	}

	// Should contain executable "local-script" from vpn dir
	if !contains(results, "local-script") {
		t.Errorf("results = %v, should contain 'local-script'", results)
	}
}

func TestCompleteCommands_WithCommandsPath(t *testing.T) {
	cfgPath, _ := writeTestConfig(t)
	app, _, _ := newTestApp(t, cfgPath)

	results, _ := completeCommands(app, "acme", "vpn", "")

	// Should contain "test-cmd" from commands_path
	if !contains(results, "test-cmd") {
		t.Errorf("results = %v, should contain 'test-cmd' from commands_path", results)
	}
}

func TestCompleteCommands_PathFallback(t *testing.T) {
	cfgPath, dir := writeTestConfig(t)
	app, _, _ := newTestApp(t, cfgPath)

	// Create a subdir in vpnDir
	vpnDir := filepath.Join(dir, "vpn")
	os.MkdirAll(filepath.Join(vpnDir, "scripts"), 0755)
	os.WriteFile(filepath.Join(vpnDir, "scripts", "deploy.sh"), []byte("#!/bin/sh\n"), 0755)

	// "./" prefix should switch to file listing
	results, _ := completeCommands(app, "acme", "vpn", "./")
	if !contains(results, "./scripts/") {
		t.Errorf("results = %v, should contain './scripts/' for path-like input", results)
	}
	if !contains(results, "./local-script") {
		t.Errorf("results = %v, should contain './local-script'", results)
	}

	// Deeper path
	results2, _ := completeCommands(app, "acme", "vpn", "./scripts/")
	if !contains(results2, "./scripts/deploy.sh") {
		t.Errorf("results = %v, should contain './scripts/deploy.sh'", results2)
	}
}

func TestCompleteCommands_UnknownDomain(t *testing.T) {
	cfgPath, _ := writeTestConfig(t)
	app, _, _ := newTestApp(t, cfgPath)

	results, _ := completeCommands(app, "acme", "nonexistent", "")
	// Should still return builtins even if domain is wrong
	// Actually computeEffectiveForCompletion will return nil, but builtins are still added
	if contains(results, "up") {
		t.Errorf("results = %v, should not contain 'up' for unknown domain", results)
	}
}

func TestCompleteArgs(t *testing.T) {
	cfgPath, dir := writeTestConfig(t)
	app, _, _ := newTestApp(t, cfgPath)

	// Create some files in vpnDir for completion
	vpnDir := filepath.Join(dir, "vpn")
	os.WriteFile(filepath.Join(vpnDir, "deploy.yml"), []byte(""), 0644)
	os.MkdirAll(filepath.Join(vpnDir, "inventory"), 0755)

	results, _ := completeArgs(app, "acme", "vpn", "")

	// Should contain files from vpn dir
	if !contains(results, "deploy.yml") {
		t.Errorf("results = %v, should contain 'deploy.yml'", results)
	}
	if !contains(results, "inventory/") {
		t.Errorf("results = %v, should contain 'inventory/'", results)
	}
}

func TestCompleteArgs_WithPrefix(t *testing.T) {
	cfgPath, dir := writeTestConfig(t)
	app, _, _ := newTestApp(t, cfgPath)

	vpnDir := filepath.Join(dir, "vpn")
	os.WriteFile(filepath.Join(vpnDir, "deploy.yml"), []byte(""), 0644)
	os.WriteFile(filepath.Join(vpnDir, "status.sh"), []byte(""), 0644)

	results, _ := completeArgs(app, "acme", "vpn", "de")

	if !contains(results, "deploy.yml") {
		t.Errorf("results = %v, should contain 'deploy.yml'", results)
	}
	if contains(results, "status.sh") {
		t.Errorf("results = %v, should not contain 'status.sh' (wrong prefix)", results)
	}
}

func TestCompleteArgs_Subdirectory(t *testing.T) {
	cfgPath, dir := writeTestConfig(t)
	app, _, _ := newTestApp(t, cfgPath)

	vpnDir := filepath.Join(dir, "vpn")
	os.MkdirAll(filepath.Join(vpnDir, "inventory"), 0755)
	os.WriteFile(filepath.Join(vpnDir, "inventory", "prod.ini"), []byte(""), 0644)
	os.WriteFile(filepath.Join(vpnDir, "inventory", "staging.ini"), []byte(""), 0644)

	results, _ := completeArgs(app, "acme", "vpn", "inventory/")

	if !contains(results, "inventory/prod.ini") {
		t.Errorf("results = %v, should contain 'inventory/prod.ini'", results)
	}
	if !contains(results, "inventory/staging.ini") {
		t.Errorf("results = %v, should contain 'inventory/staging.ini'", results)
	}
}

func TestListExecutables(t *testing.T) {
	dir := t.TempDir()

	// Executable file
	os.WriteFile(filepath.Join(dir, "run.sh"), []byte("#!/bin/sh\n"), 0755)
	// Non-executable file
	os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("text"), 0644)
	// Directory
	os.MkdirAll(filepath.Join(dir, "subdir"), 0755)

	names := listExecutables(dir)
	if !contains(names, "run.sh") {
		t.Errorf("names = %v, should contain 'run.sh'", names)
	}
	if contains(names, "readme.txt") {
		t.Errorf("names = %v, should not contain 'readme.txt'", names)
	}
	if contains(names, "subdir") {
		t.Errorf("names = %v, should not contain 'subdir'", names)
	}
}

func TestListExecutables_NonexistentDir(t *testing.T) {
	names := listExecutables("/nonexistent-dir-xyz-999")
	if len(names) != 0 {
		t.Errorf("names = %v, want empty for nonexistent dir", names)
	}
}

func TestListFiles(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "file1.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(dir, "file2.yml"), []byte(""), 0644)
	os.WriteFile(filepath.Join(dir, ".hidden"), []byte(""), 0644)
	os.MkdirAll(filepath.Join(dir, "subdir"), 0755)

	results := listFiles(dir, "")
	sort.Strings(results)

	if !contains(results, "file1.txt") {
		t.Errorf("results = %v, should contain 'file1.txt'", results)
	}
	if !contains(results, "file2.yml") {
		t.Errorf("results = %v, should contain 'file2.yml'", results)
	}
	if !contains(results, "subdir/") {
		t.Errorf("results = %v, should contain 'subdir/'", results)
	}
	// Hidden files should be excluded by default
	if contains(results, ".hidden") {
		t.Errorf("results = %v, should not contain '.hidden'", results)
	}
}

func TestListFiles_HiddenPrefix(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, ".env"), []byte(""), 0644)
	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(""), 0644)

	results := listFiles(dir, ".")
	if !contains(results, ".env") {
		t.Errorf("results = %v, should contain '.env' when prefix is '.'", results)
	}
	if !contains(results, ".gitignore") {
		t.Errorf("results = %v, should contain '.gitignore' when prefix is '.'", results)
	}
}

func TestListFiles_WithPrefix(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "deploy.yml"), []byte(""), 0644)
	os.WriteFile(filepath.Join(dir, "debug.log"), []byte(""), 0644)
	os.WriteFile(filepath.Join(dir, "status.sh"), []byte(""), 0644)

	results := listFiles(dir, "de")
	if !contains(results, "deploy.yml") {
		t.Errorf("results = %v, should contain 'deploy.yml'", results)
	}
	if !contains(results, "debug.log") {
		t.Errorf("results = %v, should contain 'debug.log'", results)
	}
	if contains(results, "status.sh") {
		t.Errorf("results = %v, should not contain 'status.sh'", results)
	}
}

func TestComputeEffectiveForCompletion(t *testing.T) {
	cfgPath, _ := writeTestConfig(t)
	app, _, _ := newTestApp(t, cfgPath)

	eff := computeEffectiveForCompletion(app, "acme", "vpn")
	if eff == nil {
		t.Fatal("computeEffectiveForCompletion returned nil")
	}
	if eff.Dir == "" {
		t.Error("eff.Dir should not be empty")
	}
	if len(eff.Profiles) == 0 {
		t.Error("eff.Profiles should not be empty")
	}
}

func TestComputeEffectiveForCompletion_UnknownClient(t *testing.T) {
	cfgPath, _ := writeTestConfig(t)
	app, _, _ := newTestApp(t, cfgPath)

	eff := computeEffectiveForCompletion(app, "nonexistent", "vpn")
	if eff != nil {
		t.Error("expected nil for unknown client")
	}
}

func TestComputeEffectiveForCompletion_UnknownDomain(t *testing.T) {
	cfgPath, _ := writeTestConfig(t)
	app, _, _ := newTestApp(t, cfgPath)

	eff := computeEffectiveForCompletion(app, "acme", "nonexistent")
	if eff != nil {
		t.Error("expected nil for unknown domain")
	}
}

func TestComputeEffectiveForCompletion_NilConfig(t *testing.T) {
	app := NewApp("test")
	eff := computeEffectiveForCompletion(app, "acme", "vpn")
	if eff != nil {
		t.Error("expected nil for nil config")
	}
}

func contains(s []string, val string) bool {
	for _, v := range s {
		if v == val {
			return true
		}
	}
	return false
}
