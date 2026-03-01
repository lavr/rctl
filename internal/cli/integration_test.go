package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lavr/rctl/internal/config"
)

func writeTestConfig(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()

	// Create a commands directory with a test script
	cmdDir := filepath.Join(dir, "commands")
	os.MkdirAll(cmdDir, 0755)
	os.WriteFile(filepath.Join(cmdDir, "test-cmd"), []byte("#!/bin/sh\necho test-cmd-output\n"), 0755)

	// Create domain directories
	vpnDir := filepath.Join(dir, "vpn")
	ansibleDir := filepath.Join(dir, "ansible")
	os.MkdirAll(vpnDir, 0755)
	os.MkdirAll(ansibleDir, 0755)

	// Write a test executable in vpnDir
	os.WriteFile(filepath.Join(vpnDir, "local-script"), []byte("#!/bin/sh\necho local-script-output\n"), 0755)

	cfgPath := filepath.Join(dir, "config.yaml")
	cfg := `
version: 1
defaults:
  env:
    LANG: "C"
  commands_path:
    - "` + cmdDir + `"
builtins:
  kctx:
    cmd: "echo"
    args: ["kubectx-builtin"]
clients:
  acme:
    aliases: ["ac"]
    tags: ["prod"]
    defaults:
      env:
        RCTL_CLIENT: "acme"
    domains:
      vpn:
        dir: "` + vpnDir + `"
        env:
          VPN_ENV: "prod"
        default_args: ["--verbose"]
        profiles:
          up:
            cmd: "echo"
            args: ["vpn-up"]
        aliases: ["v"]
      vpn-staging:
        extends: "vpn"
        env:
          VPN_ENV: "staging"
      ansible:
        dir: "` + ansibleDir + `"
        env:
          ANSIBLE_COLOR: "1"
        profiles:
          ping:
            cmd: "echo"
            args: ["ansible-ping"]
`
	os.WriteFile(cfgPath, []byte(cfg), 0644)
	return cfgPath, dir
}

func newTestApp(t *testing.T, cfgPath string) (*App, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	app := NewApp("test")
	app.ConfigPath = cfgPath

	var out, errOut bytes.Buffer
	app.Out = &out
	app.Err = &errOut

	if err := app.LoadConfig(); err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	app.SetupVerbose()
	return app, &out, &errOut
}

func TestIntegration_ClientsList(t *testing.T) {
	cfgPath, _ := writeTestConfig(t)
	app, out, _ := newTestApp(t, cfgPath)

	clients := app.Cfg.Clients
	if len(clients) != 1 {
		t.Fatalf("clients = %d, want 1", len(clients))
	}

	// Test that clients command would list acme
	names := make([]string, 0)
	for name := range clients {
		names = append(names, name)
	}
	if names[0] != "acme" {
		t.Errorf("client = %q, want acme", names[0])
	}
	_ = out
}

func TestIntegration_DryRun(t *testing.T) {
	cfgPath, _ := writeTestConfig(t)
	app, out, _ := newTestApp(t, cfgPath)
	app.DryRun = true

	err := runDispatch(app, "acme", "vpn", "up", nil)
	if err != nil {
		t.Fatalf("runDispatch: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "cwd:") {
		t.Errorf("output should contain cwd, got: %s", output)
	}
	if !strings.Contains(output, "echo") {
		t.Errorf("output should contain echo, got: %s", output)
	}
	if !strings.Contains(output, "vpn-up") {
		t.Errorf("output should contain vpn-up, got: %s", output)
	}
}

func TestIntegration_AliasDispatch(t *testing.T) {
	cfgPath, _ := writeTestConfig(t)
	app, out, _ := newTestApp(t, cfgPath)
	app.DryRun = true

	// Use client alias "ac" and domain alias "v"
	err := runDispatch(app, "ac", "v", "up", nil)
	if err != nil {
		t.Fatalf("runDispatch: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "vpn-up") {
		t.Errorf("output should contain vpn-up, got: %s", output)
	}
}

func TestIntegration_ExtendsDispatch(t *testing.T) {
	cfgPath, _ := writeTestConfig(t)
	app, out, _ := newTestApp(t, cfgPath)
	app.DryRun = true

	err := runDispatch(app, "acme", "vpn-staging", "up", nil)
	if err != nil {
		t.Fatalf("runDispatch: %v", err)
	}

	output := out.String()
	// Should inherit the profile from vpn
	if !strings.Contains(output, "vpn-up") {
		t.Errorf("output should contain vpn-up (inherited), got: %s", output)
	}
	// Should have staging env
	if !strings.Contains(output, "VPN_ENV=staging") {
		t.Errorf("output should contain VPN_ENV=staging, got: %s", output)
	}
}

func TestIntegration_PrintCmdline(t *testing.T) {
	cfgPath, _ := writeTestConfig(t)
	app, out, _ := newTestApp(t, cfgPath)
	app.PrintCmdline = true
	app.DryRun = true

	err := runDispatch(app, "acme", "vpn", "up", nil)
	if err != nil {
		t.Fatalf("runDispatch: %v", err)
	}

	output := strings.TrimSpace(out.String())
	if !strings.Contains(output, "echo") {
		t.Errorf("output should contain the command, got: %s", output)
	}
}

func TestIntegration_PrintEnv(t *testing.T) {
	cfgPath, _ := writeTestConfig(t)
	app, out, _ := newTestApp(t, cfgPath)
	app.PrintEnv = true
	app.DryRun = true

	err := runDispatch(app, "acme", "vpn", "up", nil)
	if err != nil {
		t.Fatalf("runDispatch: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "VPN_ENV=prod") {
		t.Errorf("output should contain VPN_ENV=prod, got: %s", output)
	}
}

func TestIntegration_CommandNotFound(t *testing.T) {
	cfgPath, _ := writeTestConfig(t)
	app, _, errOut := newTestApp(t, cfgPath)

	err := runDispatch(app, "acme", "vpn", "nonexistent-cmd-xyz-999", nil)
	if err != nil {
		t.Fatalf("runDispatch should not return error: %v", err)
	}

	if exitCode != 127 {
		t.Errorf("exit code = %d, want 127", exitCode)
	}

	output := errOut.String()
	if !strings.Contains(output, "not found") {
		t.Errorf("stderr should contain 'not found', got: %s", output)
	}
}

func TestIntegration_BuiltinCommand(t *testing.T) {
	cfgPath, _ := writeTestConfig(t)
	app, out, _ := newTestApp(t, cfgPath)
	app.DryRun = true

	err := runDispatch(app, "acme", "vpn", "kctx", nil)
	if err != nil {
		t.Fatalf("runDispatch: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "echo") {
		t.Errorf("output should resolve kctx to echo, got: %s", output)
	}
}

func TestIntegration_ShowJSON(t *testing.T) {
	cfgPath, _ := writeTestConfig(t)
	app, _, _ := newTestApp(t, cfgPath)

	// Simulate show --json
	var out bytes.Buffer
	app.Out = &out

	client := app.Cfg.Clients["acme"]
	eff := config.ComputeEffective(app.Cfg.Defaults, client.Defaults, client.Domains["vpn"].DomainConfig)

	err := printEffectiveJSON(app, "acme", "vpn", eff)
	if err != nil {
		t.Fatalf("printEffectiveJSON: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out.String())
	}

	if result["client"] != "acme" {
		t.Errorf("client = %v, want acme", result["client"])
	}
	if result["domain"] != "vpn" {
		t.Errorf("domain = %v, want vpn", result["domain"])
	}
}

func TestIntegration_RealExecution(t *testing.T) {
	cfgPath, _ := writeTestConfig(t)
	app, out, _ := newTestApp(t, cfgPath)

	// Actually run the echo command (profile "up" maps to "echo vpn-up")
	err := runDispatch(app, "acme", "vpn", "up", []string{"extra-arg"})
	if err != nil {
		t.Fatalf("runDispatch: %v", err)
	}

	output := strings.TrimSpace(out.String())
	// echo receives: default_args + profile.args + user args
	// = --verbose vpn-up extra-arg
	if !strings.Contains(output, "--verbose vpn-up extra-arg") {
		t.Errorf("output = %q, want to contain '--verbose vpn-up extra-arg'", output)
	}
}

func TestIntegration_CwdExecution(t *testing.T) {
	cfgPath, _ := writeTestConfig(t)
	app, out, _ := newTestApp(t, cfgPath)

	// Run local-script which exists in vpnDir (cwd for vpn domain)
	err := runDispatch(app, "acme", "vpn", "local-script", nil)
	if err != nil {
		t.Fatalf("runDispatch: %v", err)
	}

	output := strings.TrimSpace(out.String())
	// The script outputs "local-script-output", but default_args are passed first
	// Actually echo local-script would print differently...
	// local-script is a shell script that echoes "local-script-output"
	// default_args are passed as args to local-script, which ignores them
	if !strings.Contains(output, "local-script-output") {
		t.Errorf("output = %q, want to contain 'local-script-output'", output)
	}
}
