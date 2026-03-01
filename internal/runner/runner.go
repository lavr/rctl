package runner

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// Params holds everything needed to run a command.
type Params struct {
	ExecPath string
	Args     []string // arguments (not including the executable itself)
	Dir      string
	Env      map[string]string
	PathAdd  []string

	DryRun       bool
	PrintCmdline bool
	PrintEnv     bool

	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader
}

// Run executes the command with the given params.
// Returns the exit code of the child process.
func Run(p Params) int {
	// Build environment
	envOverrides := make(map[string]string, len(p.Env)+1)
	for k, v := range p.Env {
		envOverrides[k] = v
	}
	if len(p.PathAdd) > 0 {
		envOverrides["PATH"] = PrependPath(p.PathAdd)
	}
	env := BuildEnv(os.Environ(), envOverrides)

	// Handle output modes
	if p.PrintCmdline {
		cmdline := shellEscape(p.ExecPath, p.Args)
		fmt.Fprintln(p.Stdout, cmdline)
		if !p.DryRun {
			return execute(p, env)
		}
		return 0
	}

	if p.PrintEnv {
		for _, e := range env {
			if k, _, ok := strings.Cut(e, "="); ok {
				if _, overridden := envOverrides[k]; overridden {
					fmt.Fprintln(p.Stdout, e)
				}
			}
		}
		if !p.DryRun {
			return execute(p, env)
		}
		return 0
	}

	if p.DryRun {
		fmt.Fprintf(p.Stdout, "cwd: %s\n", p.Dir)
		fmt.Fprintf(p.Stdout, "cmd: %s\n", shellEscape(p.ExecPath, p.Args))
		if len(envOverrides) > 0 {
			fmt.Fprintf(p.Stdout, "env:\n")
			for k, v := range envOverrides {
				fmt.Fprintf(p.Stdout, "  %s=%s\n", k, v)
			}
		}
		return 0
	}

	return execute(p, env)
}

func execute(p Params, env []string) int {
	cmd := exec.Command(p.ExecPath, p.Args...)
	cmd.Dir = p.Dir
	cmd.Env = env
	cmd.Stdout = p.Stdout
	cmd.Stderr = p.Stderr
	cmd.Stdin = p.Stdin

	cancel := forwardSignals(cmd)
	defer cancel()

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus()
			}
			return 1
		}
		fmt.Fprintf(p.Stderr, "exec error: %v\n", err)
		return 1
	}
	return 0
}

func shellEscape(execPath string, args []string) string {
	parts := make([]string, 0, 1+len(args))
	parts = append(parts, quote(execPath))
	for _, a := range args {
		parts = append(parts, quote(a))
	}
	return strings.Join(parts, " ")
}

func quote(s string) string {
	if s == "" {
		return "''"
	}
	safe := true
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') ||
			c == '-' || c == '_' || c == '.' || c == '/' || c == ':' || c == '=' || c == ',' || c == '+') {
			safe = false
			break
		}
	}
	if safe {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}
