package runner

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_DryRun(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run(Params{
		ExecPath: "/usr/bin/echo",
		Args:     []string{"hello", "world"},
		Dir:      "/tmp",
		Env:      map[string]string{"FOO": "bar"},
		DryRun:   true,
		Stdout:   &out,
		Stderr:   &errOut,
	})

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}

	output := out.String()
	if !strings.Contains(output, "cwd: /tmp") {
		t.Errorf("output missing cwd, got: %s", output)
	}
	if !strings.Contains(output, "/usr/bin/echo") {
		t.Errorf("output missing command, got: %s", output)
	}
	if !strings.Contains(output, "FOO=bar") {
		t.Errorf("output missing env, got: %s", output)
	}
}

func TestRun_ExecEcho(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run(Params{
		ExecPath: "/bin/echo",
		Args:     []string{"hello"},
		Dir:      os.TempDir(),
		Stdout:   &out,
		Stderr:   &errOut,
		Stdin:    strings.NewReader(""),
	})

	if code != 0 {
		t.Errorf("exit code = %d, want 0, stderr: %s", code, errOut.String())
	}
	if strings.TrimSpace(out.String()) != "hello" {
		t.Errorf("output = %q, want %q", strings.TrimSpace(out.String()), "hello")
	}
}

func TestRun_ExitCode(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	// Create a script that exits with code 42
	dir := t.TempDir()
	script := filepath.Join(dir, "exit42.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nexit 42\n"), 0755); err != nil {
		t.Fatal(err)
	}

	code := Run(Params{
		ExecPath: script,
		Dir:      dir,
		Stdout:   &out,
		Stderr:   &errOut,
		Stdin:    strings.NewReader(""),
	})

	if code != 42 {
		t.Errorf("exit code = %d, want 42", code)
	}
}

func TestRun_EnvOverride(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	// Create a script that prints the CUSTOM_VAR env
	dir := t.TempDir()
	script := filepath.Join(dir, "printenv.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho $CUSTOM_VAR\n"), 0755); err != nil {
		t.Fatal(err)
	}

	code := Run(Params{
		ExecPath: script,
		Dir:      dir,
		Env:      map[string]string{"CUSTOM_VAR": "hello_from_rctl"},
		Stdout:   &out,
		Stderr:   &errOut,
		Stdin:    strings.NewReader(""),
	})

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if strings.TrimSpace(out.String()) != "hello_from_rctl" {
		t.Errorf("output = %q, want %q", strings.TrimSpace(out.String()), "hello_from_rctl")
	}
}

func TestRun_PrintCmdline(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run(Params{
		ExecPath:     "/usr/bin/echo",
		Args:         []string{"hello", "world"},
		Dir:          "/tmp",
		PrintCmdline: true,
		DryRun:       true,
		Stdout:       &out,
		Stderr:       &errOut,
	})

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	output := strings.TrimSpace(out.String())
	if !strings.Contains(output, "/usr/bin/echo") {
		t.Errorf("output = %q, should contain /usr/bin/echo", output)
	}
}

func TestRun_PrintEnv(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run(Params{
		ExecPath: "/usr/bin/echo",
		Dir:      "/tmp",
		Env:      map[string]string{"MY_KEY": "my_val"},
		PrintEnv: true,
		DryRun:   true,
		Stdout:   &out,
		Stderr:   &errOut,
	})

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	output := out.String()
	if !strings.Contains(output, "MY_KEY=my_val") {
		t.Errorf("output = %q, should contain MY_KEY=my_val", output)
	}
}
