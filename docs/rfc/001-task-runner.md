# RFC-001: Task Runner Fallback

## Status

Implemented (2026-03-02)

## Motivation

Many projects have their working commands described in a Makefile, Taskfile.yml, or Justfile. Currently, to invoke a make target through rctl you need to write:

```
rctl client1 vpn make deploy
```

It would be nicer to have it shorter:

```
rctl client1 vpn deploy
```

Regular commands (`ls`, `git`, `ansible-playbook`) should continue to work as before.

## Design

### Config

New `task_runner` field in `DomainConfig`:

```yaml
domains:
  vpn:
    dir: "~/work/client1/vpn"
    task_runner: auto  # "auto" | "make" | "task" | "just" | ""
```

- `""` (default) — task runner disabled, behavior unchanged
- `auto` — auto-detect based on file presence in `eff.Dir`
- `make` / `task` / `just` — use a specific runner

### Auto-detection (`auto`)

Files in `eff.Dir` are checked in priority order:

1. `Justfile` → `just`
2. `Taskfile.yml` or `Taskfile.yaml` → `task`
3. `Makefile` or `GNUmakefile` → `make`

The first file found determines the runner. If none are found — the task runner is not used (no error).

### Command resolution chain (change)

Current:

```
1. profiles
2. builtins
3. findExec (cwd → commands_path → PATH)
4. error: command not found
```

New:

```
1. profiles
2. builtins
3. findExec (cwd → commands_path → PATH)
4. task runner (if configured)
5. error: command not found
```

Task runner is a **fallback**. If the command is found as an executable, it runs directly. Only when the command is not found anywhere does rctl delegate to the task runner.

### Task runner resolution mechanics

When a command is not found and the task runner is active:

1. Determine runner (`auto` → check files, or take from config)
2. Find the runner executable via `exec.LookPath`
3. Return `Result{ExecPath: runnerPath, Argv: [cmdName, ...userArgs], Source: "task_runner"}`

Target validation **is not performed** — the runner itself will produce an error if the target does not exist.

### Examples

Config:

```yaml
clients:
  client1:
    domains:
      vpn:
        dir: "~/work/client1/vpn"
        task_runner: auto
```

In `~/work/client1/vpn/Makefile`:

```makefile
deploy:
	ansible-playbook deploy.yml

status:
	./vpnctl status
```

Usage:

```
rctl client1 vpn deploy           → make deploy         (source: task_runner)
rctl client1 vpn status           → make status         (source: task_runner)
rctl client1 vpn ls               → /bin/ls             (source: PATH, ls found)
rctl client1 vpn make deploy      → make deploy         (source: PATH, make found)
rctl client1 vpn git status       → git status          (source: PATH)
rctl client1 vpn nonexistent      → make nonexistent    → make: *** No rule... (error from make)
```

### --dry-run output

```
cwd: /Users/user/work/client1/vpn
cmd: /usr/bin/make deploy
source: task_runner (auto → Makefile)
```

## Code changes

### 1. `internal/config/types.go`

Add field to `DomainConfig`:

```go
type DomainConfig struct {
    // ... existing fields ...
    TaskRunner string `yaml:"task_runner,omitempty"`
}
```

Add field to `EffectiveConfig`:

```go
type EffectiveConfig struct {
    // ... existing fields ...
    TaskRunner string
}
```

### 2. `internal/config/effective.go`

In `ComputeEffective` — merge `TaskRunner` (last non-empty wins):

```go
if layer.TaskRunner != "" {
    eff.TaskRunner = layer.TaskRunner
}
```

### 3. `internal/resolve/command.go`

Modify `ResolveCommand` — add fallback after `findExec`:

```go
execPath, foundSource := findExec(realCmd, eff, verbose)
if execPath == "" {
    // Task runner fallback
    if runner := detectTaskRunner(eff, verbose); runner != "" {
        runnerPath, err := exec.LookPath(runner)
        if err == nil {
            verbose.Printf("delegating %q to task runner %s", cmdName, runner)
            argv = append([]string{cmdName}, userArgs...)
            return &Result{
                ExecPath: runnerPath,
                Argv:     argv,
                Source:   "task_runner",
            }, nil
        }
    }
    return nil, fmt.Errorf("command %q not found", realCmd)
}
```

New function `detectTaskRunner`:

```go
func detectTaskRunner(eff *config.EffectiveConfig, verbose *log.Logger) string {
    runner := eff.TaskRunner
    if runner == "" {
        return ""
    }
    if runner != "auto" {
        return runner  // "make", "task", "just"
    }

    // Auto-detect
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
```

### 4. `internal/cli/show.go`

Display `task_runner` in `show` output if set.

### 5. `internal/cli/doctor.go`

Add check: if `task_runner` is set explicitly (not auto), verify that the runner is available in PATH.

## Out of scope

- Target auto-completion (shell completion) — can be added later
- Passing arguments to the runner (e.g. `make -j4`) — use `default_args`
- Runner cascade (try just, then make) — use `auto`
- Target existence validation before execution

## Backward compatibility

Full. By default `task_runner` is empty — behavior is unchanged. The new field is optional.
