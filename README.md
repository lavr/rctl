# rctl

CLI tool for running commands in a client + domain context.

If you work with multiple clients/projects and each has its own VPN, Ansible, Kubernetes, and other tools — `rctl` lets you switch between them with a single command, automatically setting the working directory, environment variables, and arguments.

## Problem

A typical workday:

```bash
cd ~/work/client-a/vpn && VPN_PROFILE=main wg-quick up wg0
cd ~/work/client-a/ansible && ANSIBLE_FORCE_COLOR=1 ansible all -m ping
cd ~/work/client-b/k8s && KUBECONFIG=./kubeconfig.yaml kubectl apply -f .
```

Remembering paths, variables, and arguments for each client is tedious and error-prone.

## Solution

Describe everything once in a YAML config:

```yaml
version: 1

clients:
  client-a:
    aliases: ["a"]
    domains:
      vpn:
        dir: "~/work/client-a/vpn"
        env: { VPN_PROFILE: "main" }
        profiles:
          up:   { cmd: "wg-quick", args: ["up", "wg0"] }
          down: { cmd: "wg-quick", args: ["down", "wg0"] }
      ansible:
        dir: "~/work/client-a/ansible"
        env: { ANSIBLE_FORCE_COLOR: "1" }
        profiles:
          ping: { cmd: "ansible", args: ["all", "-m", "ping"] }
```

Then just work:

```bash
rctl client-a vpn up           # wg-quick up wg0 in ~/work/client-a/vpn
rctl client-a ansible ping     # ansible all -m ping
rctl a vpn down                # via alias
```

`rctl` handles `chdir`, sets env variables, substitutes arguments, and forwards the exit code.

## Features

- **Clients and domains** — structured context for commands
- **Profiles** — named commands with preset arguments
- **3-layer merge** — defaults → client.defaults → domain (maps override, lists concat)
- **Inheritance** — `extends` for domains, configuration reuse
- **Aliases** — short names for clients and domains
- **Templates** — Go `text/template` in env, args, and paths (`{{ .Client.Name }}`, `{{ .Vars.region }}`)
- **Includes** — split config across files with glob support
- **Secrets** — env and command providers (`{{ secret "command:op" "path" }}`)
- **Policies** — allow/deny command lists, path restrictions, require_tty
- **Audit** — JSON Lines log with automatic redaction of sensitive data
- **Diagnostics** — `rctl doctor` checks config, paths, aliases, templates

## Installation

### Homebrew

```bash
brew install lavr/tap/rctl
```

### Go

```bash
go install github.com/lavr/rctl/cmd/rctl@latest
```

### From source

```bash
git clone https://github.com/lavr/rctl.git
cd rctl
make build
```

## Quick start

Create `~/.config/rctl/config.yaml` and start working.

See [docs/Quickstart.md](docs/Quickstart.md) for details.

## CLI

```
rctl <client> <domain> <cmd> [args...]    main mode
rctl <client> <domain> -- <cmd> [args...] explicit argument boundary

rctl clients                              list clients
rctl domains <client>                     list domains
rctl show <client> <domain> [--json]      effective config
rctl which <client> <domain> <cmd> [--json] path to executable
rctl policy <client> <domain> <cmd>       policy check
rctl doctor                               config diagnostics
rctl version                              version
```

Global flags:

| Flag | Description |
|------|-------------|
| `--config <path>` | Path to config file |
| `--verbose` | Verbose logging to stderr |
| `--dry-run` | Show what would be executed without running |
| `--print-cmdline` | Print the command line |
| `--print-env` | Print env overrides |

## Config structure

```yaml
version: 1
includes: [...]          # additional files (glob)
defaults:                 # global settings
  env: {}
  commands_path: []
  policy: { ... }
vars: {}                  # template variables
builtins: {}              # global command aliases
secrets: { providers: {} }
clients:
  <name>:
    aliases: []
    defaults: { ... }    # client settings
    domains:
      <name>:
        dir: ""
        env: {}
        default_args: []
        profiles: {}
        extends: ""       # inherit from another domain
        aliases: []
```

Full example: [examples/config.yaml](examples/config.yaml).

## Examples

```bash
# Run via profile
rctl clientX vpn up

# Run an arbitrary command in domain context
rctl clientX ansible ansible-playbook deploy.yml

# Via alias
rctl inv ans ping

# Debugging
rctl clientX vpn --dry-run up
rctl which clientX vpn wg
rctl doctor
```

## License

MIT
