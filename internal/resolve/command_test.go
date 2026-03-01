package resolve

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/lavr/rctl/internal/config"
)

var nopLogger = log.New(io.Discard, "", 0)

func writeExec(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}
}

func TestResolveCommand_Profile(t *testing.T) {
	dir := t.TempDir()
	writeExec(t, dir, "wg-up")

	eff := &config.EffectiveConfig{
		Profiles: map[string]config.Profile{
			"up": {Cmd: "wg-up", Args: []string{"--profile", "main"}},
		},
		CommandsPath: []string{dir},
		Env:          map[string]string{},
		Vars:         map[string]string{},
	}

	result, err := ResolveCommand("up", []string{"extra"}, eff, nil, nopLogger)
	if err != nil {
		t.Fatalf("ResolveCommand failed: %v", err)
	}

	if result.Source != "profile" {
		t.Errorf("source = %q, want %q", result.Source, "profile")
	}
	if filepath.Base(result.ExecPath) != "wg-up" {
		t.Errorf("exec = %q, want wg-up", result.ExecPath)
	}
	// Args: default_args + profile.args + user args
	wantArgs := []string{"--profile", "main", "extra"}
	if len(result.Argv) != len(wantArgs) {
		t.Fatalf("argv = %v, want %v", result.Argv, wantArgs)
	}
	for i, v := range wantArgs {
		if result.Argv[i] != v {
			t.Errorf("argv[%d] = %q, want %q", i, result.Argv[i], v)
		}
	}
}

func TestResolveCommand_Builtin(t *testing.T) {
	dir := t.TempDir()
	writeExec(t, dir, "kubectx")

	eff := &config.EffectiveConfig{
		CommandsPath: []string{dir},
		Profiles:     map[string]config.Profile{},
		Env:          map[string]string{},
		Vars:         map[string]string{},
	}
	builtins := map[string]config.Builtin{
		"kctx": {Cmd: "kubectx"},
	}

	result, err := ResolveCommand("kctx", nil, eff, builtins, nopLogger)
	if err != nil {
		t.Fatalf("ResolveCommand failed: %v", err)
	}
	if result.Source != "builtin" {
		t.Errorf("source = %q, want %q", result.Source, "builtin")
	}
}

func TestResolveCommand_CWD(t *testing.T) {
	dir := t.TempDir()
	writeExec(t, dir, "my-script")

	eff := &config.EffectiveConfig{
		Dir:      dir,
		Profiles: map[string]config.Profile{},
		Env:      map[string]string{},
		Vars:     map[string]string{},
	}

	result, err := ResolveCommand("my-script", nil, eff, nil, nopLogger)
	if err != nil {
		t.Fatalf("ResolveCommand failed: %v", err)
	}
	if result.Source != "cwd" {
		t.Errorf("source = %q, want %q", result.Source, "cwd")
	}
}

func TestResolveCommand_CommandsPath(t *testing.T) {
	cmdDir := t.TempDir()
	writeExec(t, cmdDir, "special-cmd")

	eff := &config.EffectiveConfig{
		Dir:          t.TempDir(), // different dir, no executable here
		CommandsPath: []string{cmdDir},
		Profiles:     map[string]config.Profile{},
		Env:          map[string]string{},
		Vars:         map[string]string{},
	}

	result, err := ResolveCommand("special-cmd", nil, eff, nil, nopLogger)
	if err != nil {
		t.Fatalf("ResolveCommand failed: %v", err)
	}
	if result.Source != "commands_path" {
		t.Errorf("source = %q, want %q", result.Source, "commands_path")
	}
}

func TestResolveCommand_PATH(t *testing.T) {
	eff := &config.EffectiveConfig{
		Profiles: map[string]config.Profile{},
		Env:      map[string]string{},
		Vars:     map[string]string{},
	}

	// "echo" should be on PATH on any system
	result, err := ResolveCommand("echo", []string{"hello"}, eff, nil, nopLogger)
	if err != nil {
		t.Fatalf("ResolveCommand failed: %v", err)
	}
	if result.Source != "PATH" {
		t.Errorf("source = %q, want %q", result.Source, "PATH")
	}
}

func TestResolveCommand_NotFound(t *testing.T) {
	eff := &config.EffectiveConfig{
		Profiles: map[string]config.Profile{},
		Env:      map[string]string{},
		Vars:     map[string]string{},
	}

	_, err := ResolveCommand("nonexistent-cmd-xyz-999", nil, eff, nil, nopLogger)
	if err == nil {
		t.Error("expected error for missing command")
	}
}

func TestResolveCommand_TaskRunnerFallback(t *testing.T) {
	dir := t.TempDir()
	// Create a Makefile so auto-detect works
	os.WriteFile(filepath.Join(dir, "Makefile"), []byte("deploy:\n\techo deploy\n"), 0644)

	eff := &config.EffectiveConfig{
		Dir:        dir,
		TaskRunner: "auto",
		Profiles:   map[string]config.Profile{},
		Env:        map[string]string{},
		Vars:       map[string]string{},
	}

	// "deploy" is not an executable, so it should fall through to task runner
	result, err := ResolveCommand("deploy", []string{"prod"}, eff, nil, nopLogger)
	if err != nil {
		t.Fatalf("ResolveCommand failed: %v", err)
	}
	if result.Source != "task_runner" {
		t.Errorf("source = %q, want %q", result.Source, "task_runner")
	}
	if filepath.Base(result.ExecPath) != "make" {
		t.Errorf("exec = %q, want make", result.ExecPath)
	}
	// Argv should be: [deploy, prod] (cmdName + userArgs)
	if len(result.Argv) < 2 || result.Argv[0] != "deploy" || result.Argv[1] != "prod" {
		t.Errorf("argv = %v, want [deploy prod]", result.Argv)
	}
}

func TestResolveCommand_TaskRunnerExplicit(t *testing.T) {
	dir := t.TempDir()

	eff := &config.EffectiveConfig{
		Dir:        dir,
		TaskRunner: "make",
		Profiles:   map[string]config.Profile{},
		Env:        map[string]string{},
		Vars:       map[string]string{},
	}

	result, err := ResolveCommand("build", nil, eff, nil, nopLogger)
	if err != nil {
		t.Fatalf("ResolveCommand failed: %v", err)
	}
	if result.Source != "task_runner" {
		t.Errorf("source = %q, want %q", result.Source, "task_runner")
	}
	if result.Argv[0] != "build" {
		t.Errorf("argv[0] = %q, want %q", result.Argv[0], "build")
	}
}

func TestResolveCommand_TaskRunnerNotUsedWhenCommandFound(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "Makefile"), []byte("echo:\n\techo hi\n"), 0644)

	eff := &config.EffectiveConfig{
		Dir:        dir,
		TaskRunner: "auto",
		Profiles:   map[string]config.Profile{},
		Env:        map[string]string{},
		Vars:       map[string]string{},
	}

	// "echo" exists in PATH, so task runner should NOT be used
	result, err := ResolveCommand("echo", nil, eff, nil, nopLogger)
	if err != nil {
		t.Fatalf("ResolveCommand failed: %v", err)
	}
	if result.Source == "task_runner" {
		t.Errorf("source = %q, should not be task_runner when command is in PATH", result.Source)
	}
}

func TestResolveCommand_TaskRunnerProfileTakesPriority(t *testing.T) {
	dir := t.TempDir()
	writeExec(t, dir, "vpnctl")
	os.WriteFile(filepath.Join(dir, "Makefile"), []byte("up:\n\techo up\n"), 0644)

	eff := &config.EffectiveConfig{
		Dir:        dir,
		TaskRunner: "auto",
		Profiles: map[string]config.Profile{
			"up": {Cmd: "vpnctl", Args: []string{"up"}},
		},
		Env:  map[string]string{},
		Vars: map[string]string{},
	}

	result, err := ResolveCommand("up", nil, eff, nil, nopLogger)
	if err != nil {
		t.Fatalf("ResolveCommand failed: %v", err)
	}
	if result.Source != "profile" {
		t.Errorf("source = %q, want %q (profile should take priority over task runner)", result.Source, "profile")
	}
}

func TestResolveCommand_NoTaskRunnerNoFallback(t *testing.T) {
	eff := &config.EffectiveConfig{
		Dir:        t.TempDir(),
		TaskRunner: "",
		Profiles:   map[string]config.Profile{},
		Env:        map[string]string{},
		Vars:       map[string]string{},
	}

	_, err := ResolveCommand("nonexistent-xyz", nil, eff, nil, nopLogger)
	if err == nil {
		t.Error("expected error when task_runner is empty and command not found")
	}
}

func TestDetectTaskRunner_Auto(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		wantCmd  string
	}{
		{"Justfile", []string{"Justfile"}, "just"},
		{"Taskfile.yml", []string{"Taskfile.yml"}, "task"},
		{"Taskfile.yaml", []string{"Taskfile.yaml"}, "task"},
		{"Makefile", []string{"Makefile"}, "make"},
		{"GNUmakefile", []string{"GNUmakefile"}, "make"},
		{"Justfile wins over Makefile", []string{"Justfile", "Makefile"}, "just"},
		{"no files", []string{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, f := range tt.files {
				os.WriteFile(filepath.Join(dir, f), []byte(""), 0644)
			}

			eff := &config.EffectiveConfig{
				Dir:        dir,
				TaskRunner: "auto",
			}

			got := detectTaskRunner(eff, nopLogger)
			if got != tt.wantCmd {
				t.Errorf("detectTaskRunner() = %q, want %q", got, tt.wantCmd)
			}
		})
	}
}

func TestDetectTaskRunner_Explicit(t *testing.T) {
	eff := &config.EffectiveConfig{
		Dir:        t.TempDir(),
		TaskRunner: "just",
	}
	got := detectTaskRunner(eff, nopLogger)
	if got != "just" {
		t.Errorf("detectTaskRunner() = %q, want %q", got, "just")
	}
}

func TestDetectTaskRunner_Empty(t *testing.T) {
	eff := &config.EffectiveConfig{
		Dir:        t.TempDir(),
		TaskRunner: "",
	}
	got := detectTaskRunner(eff, nopLogger)
	if got != "" {
		t.Errorf("detectTaskRunner() = %q, want empty", got)
	}
}

func TestResolveCommand_DefaultArgs(t *testing.T) {
	dir := t.TempDir()
	writeExec(t, dir, "mycmd")

	eff := &config.EffectiveConfig{
		Dir:         dir,
		DefaultArgs: []string{"--default-flag"},
		Profiles:    map[string]config.Profile{},
		Env:         map[string]string{},
		Vars:        map[string]string{},
	}

	result, err := ResolveCommand("mycmd", []string{"user-arg"}, eff, nil, nopLogger)
	if err != nil {
		t.Fatalf("ResolveCommand failed: %v", err)
	}

	wantArgs := []string{"--default-flag", "user-arg"}
	if len(result.Argv) != len(wantArgs) {
		t.Fatalf("argv = %v, want %v", result.Argv, wantArgs)
	}
	for i, v := range wantArgs {
		if result.Argv[i] != v {
			t.Errorf("argv[%d] = %q, want %q", i, result.Argv[i], v)
		}
	}
}
