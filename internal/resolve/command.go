package resolve

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/lavr/rctl/internal/config"
)

// Result holds the resolved command information.
type Result struct {
	ExecPath string   // absolute path to the executable
	Argv     []string // full argument vector: default_args + profile.args + user args
	Source   string   // where the command was found (profile, builtin, cwd, commands_path, PATH)
}

// ResolveCommand resolves a command name through the priority chain:
// 1. profiles (domain effective)
// 2. builtins (global)
// 3. cwd (effective dir)
// 4. commands_path entries
// 5. PATH (with path_add prepended)
// 6. task runner fallback (make/task/just if configured)
//
// Returns the resolved result or an error. Exit code 127 convention for not-found
// is handled by the caller.
func ResolveCommand(cmdName string, userArgs []string, eff *config.EffectiveConfig, builtins map[string]config.Builtin, verbose *log.Logger) (*Result, error) {
	realCmd := cmdName
	var profileArgs []string
	source := "raw"

	// 1. Check profiles
	if p, ok := eff.Profiles[cmdName]; ok {
		realCmd = p.Cmd
		profileArgs = p.Args
		source = "profile"
		verbose.Printf("resolved %q via profile -> %s %v", cmdName, realCmd, profileArgs)
	} else if b, ok := builtins[cmdName]; ok {
		// 2. Check builtins
		realCmd = b.Cmd
		profileArgs = b.Args
		source = "builtin"
		verbose.Printf("resolved %q via builtin -> %s %v", cmdName, realCmd, profileArgs)
	}

	// Build argv: default_args + profile.args + user args
	var argv []string
	argv = append(argv, eff.DefaultArgs...)
	argv = append(argv, profileArgs...)
	argv = append(argv, userArgs...)

	// Find the executable
	execPath, foundSource := findExec(realCmd, eff, verbose)
	if execPath != "" {
		if source == "raw" {
			source = foundSource
		}
		return &Result{
			ExecPath: execPath,
			Argv:     argv,
			Source:   source,
		}, nil
	}

	// Task runner fallback: if command not found and task_runner is configured,
	// delegate to the task runner (make/task/just).
	if source == "raw" {
		if runner := detectTaskRunner(eff, verbose); runner != "" {
			runnerPath, err := exec.LookPath(runner)
			if err == nil {
				verbose.Printf("delegating %q to task runner %s", cmdName, runner)
				// Build argv: default_args + cmdName + user args
				var runnerArgv []string
				runnerArgv = append(runnerArgv, eff.DefaultArgs...)
				runnerArgv = append(runnerArgv, cmdName)
				runnerArgv = append(runnerArgv, userArgs...)
				return &Result{
					ExecPath: runnerPath,
					Argv:     runnerArgv,
					Source:   "task_runner",
				}, nil
			}
			verbose.Printf("task runner %q not found in PATH", runner)
		}
	}

	return nil, fmt.Errorf("command %q not found", realCmd)
}

// detectTaskRunner returns the task runner command name based on config and file detection.
// Returns empty string if no task runner should be used.
func detectTaskRunner(eff *config.EffectiveConfig, verbose *log.Logger) string {
	runner := eff.TaskRunner
	if runner == "" {
		return ""
	}
	if runner != "auto" {
		return runner
	}

	// Auto-detect by checking files in eff.Dir
	dir := eff.Dir
	if dir == "" {
		return ""
	}
	checks := []struct {
		files  []string
		runner string
	}{
		{[]string{"Justfile"}, "just"},
		{[]string{"Taskfile.yml", "Taskfile.yaml"}, "task"},
		{[]string{"Makefile", "GNUmakefile"}, "make"},
	}
	for _, c := range checks {
		for _, f := range c.files {
			if _, err := os.Stat(filepath.Join(dir, f)); err == nil {
				verbose.Printf("auto-detected task runner %q via %s", c.runner, f)
				return c.runner
			}
		}
	}
	return ""
}

// findExec searches for the executable in: cwd, commands_path, then PATH.
func findExec(name string, eff *config.EffectiveConfig, verbose *log.Logger) (string, string) {
	// If the name contains a slash, treat it as a path
	if strings.Contains(name, "/") {
		name = config.ExpandTilde(name)
		if isExecutable(name) {
			return name, "path"
		}
		// Resolve relative paths against effective dir
		if eff.Dir != "" && !filepath.IsAbs(name) {
			abs := filepath.Join(eff.Dir, name)
			if isExecutable(abs) {
				verbose.Printf("found %q in dir %s", name, eff.Dir)
				return abs, "cwd"
			}
		}
		return "", ""
	}

	// A) Check cwd (effective dir)
	if eff.Dir != "" {
		if p := findExecutable(eff.Dir, name); p != "" {
			verbose.Printf("found %q in cwd %s", name, eff.Dir)
			return p, "cwd"
		}
	}

	// B) Check commands_path
	for _, dir := range eff.CommandsPath {
		if p := findExecutable(dir, name); p != "" {
			verbose.Printf("found %q in commands_path %s", name, dir)
			return p, "commands_path"
		}
	}

	// C) Use PATH (path_add is prepended by caller via env)
	p, err := exec.LookPath(name)
	if err == nil {
		verbose.Printf("found %q in PATH -> %s", name, p)
		return p, "PATH"
	}

	verbose.Printf("command %q not found in any search path", name)
	return "", ""
}
