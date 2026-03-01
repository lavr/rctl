# RFC-002: Shell Completion

## Status

Implemented (2026-03-02)

## Motivation

Typing full client, domain, and command names manually is inconvenient. The goal is:

```
rctl li<TAB>     → rctl lifedl
rctl lifedl v<TAB> → rctl lifedl vpn
rctl lifedl vpn u<TAB> → rctl lifedl vpn up
rctl lifedl ansible ansible-playbook -i inv<TAB> → files from eff.Dir
```

## Design

### Completion levels

Positional dispatch `rctl <client> <domain> <cmd> [args...]` gives three levels:

#### Level 1: `rctl <TAB>` — clients + subcommands

Suggested:
- All client names from `cfg.Clients`
- All client aliases
- Subcommands: `clients`, `domains`, `show`, `which`, `policy`, `doctor`, `run`, `version`
- Directive `cobra.ShellCompDirectiveNoFileComp` — do not complete files

#### Level 2: `rctl <client> <TAB>` — domains

Suggested:
- All domain names for the client
- All domain aliases
- Directive `cobra.ShellCompDirectiveNoFileComp`

`resolve.FindClientByNameOrAlias` is used to determine the client. If the client is not found — empty list (no error).

#### Level 3: `rctl <client> <domain> <TAB>` — commands

Suggested:
- Profile names from effective config (profiles)
- Builtin names from `cfg.Builtins`
- Executables from `eff.Dir` (cwd)
- Executables from `commands_path`
- Everything from PATH is **not** suggested (too much noise)
- Directive `cobra.ShellCompDirectiveNoFileComp`

#### Level 4: `rctl <client> <domain> <cmd> <TAB>` — arguments

Suggested:
- Files from `eff.Dir` (not from the process CWD!)
- Directive `cobra.ShellCompDirectiveDefault` with `eff.Dir` substituted as the base directory

This is a critical point — standard file completion works from the process CWD, but we need completion from the domain's `eff.Dir`.

### Subcommands

Argument completion for subcommands:

| Subcommand | Argument 1 | Argument 2 | Argument 3 |
|------------|-----------|-----------|-----------|
| `domains` | clients | — | — |
| `show` | clients | domains | — |
| `which` | clients | domains | commands |
| `policy` | clients | domains | commands |
| `run` | clients | domains | commands |

### File completion from eff.Dir

Problem: cobra/shell completion works from CWD. We need to suggest files from `eff.Dir`.

Solution: **custom file completion**. Instead of `ShellCompDirectiveDefault` (which completes from CWD), we return a list of files from `eff.Dir` programmatically.

```go
func completeFiles(eff *config.EffectiveConfig, toComplete string) []string {
    dir := eff.Dir
    if dir == "" {
        return nil
    }
    pattern := filepath.Join(dir, toComplete+"*")
    matches, _ := filepath.Glob(pattern)
    // Return paths relative to eff.Dir
    var results []string
    for _, m := range matches {
        rel, _ := filepath.Rel(dir, m)
        if info, err := os.Stat(m); err == nil && info.IsDir() {
            rel += "/"
        }
        results = append(results, rel)
    }
    return results
}
```

### Computing effective config for completion

For levels 3-4, the effective config is needed. Computation:

1. Find client by name/alias
2. Find domain by name/alias
3. Resolve extends chain
4. `ComputeEffective(defaults, client.defaults, domain)`
5. Render templates (for `Dir`, `CommandsPath`)

**Important:** secrets are not needed during completion — the `secret` function is stubbed as a no-op.

```go
func computeEffectiveForCompletion(app *App, clientName, domainName string) *config.EffectiveConfig {
    resolvedClientName, client, err := resolve.FindClientByNameOrAlias(app.Cfg, clientName)
    if err != nil {
        return nil
    }
    resolvedDomainName, _, err := resolve.FindDomainByNameOrAlias(client, domainName)
    if err != nil {
        return nil
    }
    domainCfg, err := resolve.ResolveDomainExtends(client, resolvedDomainName)
    if err != nil {
        return nil
    }
    eff := config.ComputeEffective(app.Cfg.Defaults, client.Defaults, domainCfg)

    funcMap := config.DefaultFuncMap()
    funcMap["secret"] = func(provider, path string) (string, error) { return "", nil }
    ctx := config.NewTemplateContext(resolvedClientName, client.Tags, resolvedDomainName, eff)
    config.RenderEffective(eff, ctx, funcMap)
    return eff
}
```

### Integration with cobra

Cobra supports dynamic completion via `ValidArgsFunction` on each command. For the root command with positional dispatch:

```go
root.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    _ = preloadConfig(app) // load config if not loaded

    switch len(args) {
    case 0:
        return completeClients(app, toComplete)
    case 1:
        return completeDomains(app, args[0], toComplete)
    case 2:
        return completeCommands(app, args[0], args[1], toComplete)
    default:
        return completeArgs(app, args[0], args[1], toComplete)
    }
}
```

### Script generation

Cobra has built-in support:

```
rctl completion bash
rctl completion zsh
rctl completion fish
```

A `completion` subcommand with sub-subcommands for each shell is added. The user adds to `.bashrc` / `.zshrc`:

```bash
# bash
eval "$(rctl completion bash)"

# zsh
eval "$(rctl completion zsh)"

# fish
rctl completion fish | source
```

## Code changes

### 1. `internal/cli/completion.go` (new file)

Completion functions:
- `completeClients(app, toComplete)` — clients + aliases + subcommands
- `completeDomains(app, clientArg, toComplete)` — domains + aliases
- `completeCommands(app, clientArg, domainArg, toComplete)` — profiles + builtins + executables from dir/commands_path
- `completeArgs(app, clientArg, domainArg, toComplete)` — files from eff.Dir
- `computeEffectiveForCompletion(app, client, domain)` — compute eff without secrets
- `completeFiles(eff, toComplete)` — files relative to eff.Dir

### 2. `internal/cli/root.go`

- Add `root.ValidArgsFunction` to the root command
- Add `completionCmd(app)` subcommand for script generation

### 3. `internal/cli/completion_cmd.go` (new file)

Subcommand `rctl completion [bash|zsh|fish]` — completion script generation via cobra.

### 4. Subcommands: `domains.go`, `show.go`, `which.go`, `policy.go`, `run.go`

Add `ValidArgsFunction` for arguments:
- Argument 0 → `completeClients`
- Argument 1 → `completeDomains`
- Argument 2 → `completeCommands` (where applicable)

## Examples

```bash
$ rctl <TAB>
clients  doctor  domains  lifedl  ldl  policy  run  show  version  which

$ rctl lifedl <TAB>
ansible  ans  vpn  v

$ rctl lifedl vpn <TAB>
browse  down  list  logs  restart  status  up  vpnctl  browserctl

$ rctl lifedl ansible ansible-playbook -i <TAB>
inventory/  contrib/  run-profiles/  deploy.yml  site.yml

$ rctl show <TAB>
lifedl  ldl

$ rctl show lifedl <TAB>
ansible  ans  vpn  v
```

## Limitations and trade-offs

- **Level 3 completion does not include PATH**: too noisy, profiles + cwd + commands_path is sufficient
- **Task runner targets not included**: parsing Makefile/Taskfile/Justfile is a separate task, can be added later
- **Performance**: each `<TAB>` loads the config and computes effective. For typical configs this is <50ms
- **Errors during completion**: all errors are silently swallowed (return empty list) to avoid breaking the shell
- **`--` separator**: after `--`, completion returns files from eff.Dir

## Backward compatibility

Full. Completion is an optional feature that requires explicit activation by the user via `eval "$(rctl completion ...)"`.
