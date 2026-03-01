# Quickstart

## Installation

```bash
go install github.com/lavr/rctl/cmd/rctl@latest
```

Or from source:

```bash
git clone https://github.com/lavr/rctl.git
cd rctl
make build
# binary ./rctl is ready
```

## Minimal config

Create `~/.config/rctl/config.yaml`:

```yaml
version: 1

clients:
  mycompany:
    domains:
      vpn:
        dir: "~/work/mycompany/vpn"
        env:
          VPN_PROFILE: "main"
        profiles:
          up:
            cmd: "wg-quick"
            args: ["up", "wg0"]
          down:
            cmd: "wg-quick"
            args: ["down", "wg0"]
      ansible:
        dir: "~/work/mycompany/ansible"
        env:
          ANSIBLE_FORCE_COLOR: "1"
        profiles:
          ping:
            cmd: "ansible"
            args: ["all", "-m", "ping"]
```

## Basic commands

### Run via profile

```bash
rctl mycompany vpn up        # runs wg-quick up wg0 in ~/work/mycompany/vpn
rctl mycompany vpn down      # runs wg-quick down wg0
rctl mycompany ansible ping  # runs ansible all -m ping
```

### Run an arbitrary command

```bash
rctl mycompany vpn wg show              # runs wg show in vpn context
rctl mycompany ansible ansible-playbook deploy.yml
```

### The `--` separator

If arguments conflict with rctl flags, use `--`:

```bash
rctl mycompany vpn -- wg show --help
```

### View configuration

```bash
rctl clients                  # list clients
rctl domains mycompany        # list domains for a client
rctl show mycompany vpn       # effective config for vpn
rctl show mycompany vpn --json
```

### Debugging

```bash
rctl mycompany vpn --dry-run up       # shows cwd/cmd/env without running
rctl mycompany vpn --print-cmdline up # prints the command line
rctl mycompany vpn --print-env up     # prints env overrides
rctl which mycompany vpn wg           # shows path to executable
rctl --verbose mycompany vpn up       # verbose logging to stderr
rctl doctor                           # config diagnostics
```

## Aliases

Add aliases for quick access:

```yaml
clients:
  mycompany:
    aliases: ["mc"]
    domains:
      vpn:
        aliases: ["v"]
        # ...
```

Now you can:

```bash
rctl mc v up    # same as rctl mycompany vpn up
```

## Domain inheritance (extends)

A domain can inherit configuration from another domain:

```yaml
domains:
  vpn:
    dir: "~/work/mycompany/vpn"
    env:
      VPN_ENV: "prod"
    profiles:
      up: { cmd: "wg-quick", args: ["up", "wg0"] }
  vpn-staging:
    extends: "vpn"
    env:
      VPN_ENV: "staging"   # only overrides VPN_ENV
```

`vpn-staging` inherits `dir`, `profiles`, and all other fields from `vpn`.

## Templates

Go templates are supported in env values, args, and paths:

```yaml
domains:
  k8s:
    dir: "~/work/{{ .Client.Name }}/k8s"
    env:
      KUBECONFIG: "{{ .Dir }}/kube/{{ .Vars.region }}.yaml"
```

Available variables: `.Client.Name`, `.Client.Tags`, `.Domain.Name`, `.Dir`, `.Vars`, `.Env`, `.Host.Name`, `.Host.OS`, `.Host.Arch`, `.Time.Now`, `.Time.Date`.

## Defaults and layers

Config is merged in three layers (later wins for scalars and map keys, concat for lists):

1. `defaults` — global settings
2. `clients.<name>.defaults` — client settings
3. `clients.<name>.domains.<name>` — domain settings

```yaml
defaults:
  env:
    LANG: "C"
  commands_path:
    - "~/.local/share/rctl/commands"

clients:
  mycompany:
    defaults:
      env:
        RCTL_CLIENT: "mycompany"
    domains:
      vpn:
        env:
          VPN_ENV: "prod"
        # effective env: LANG=C, RCTL_CLIENT=mycompany, VPN_ENV=prod
```

## Includes

Split config across files:

```yaml
version: 1
includes:
  - "~/.config/rctl/clients/*.yaml"
```

Files from `~/.config/rctl/config.d/*.yaml` are loaded automatically as overrides.

## Policies

Restrict available commands:

```yaml
defaults:
  policy:
    deny_commands: ["rm", "dd"]
    allow_path_prefixes: ["/usr/local/bin/"]
```

Verify with:

```bash
rctl policy mycompany vpn wg
```

## Secrets

Connect secrets via env or external commands:

```yaml
secrets:
  providers:
    env: { enabled: true }
    command:
      enabled: true
      commands:
        op: "op read {{ .Path }}"
```

Use in templates:

```yaml
env:
  API_TOKEN: '{{ secret "env" "MY_API_TOKEN" }}'
  DB_PASS: '{{ secret "command:op" "vaults/prod/db-password" }}'
```

## Audit

Enable logging of all command executions:

```yaml
defaults:
  policy:
    audit:
      enabled: true
      sink: "file"
      file: "~/.local/state/rctl/audit.log"
      redact_env_keys: ["*_TOKEN", "*PASSWORD*", "*SECRET*"]
```

Logs are written in JSON Lines format. Sensitive values are automatically redacted.

## Environment variables

| Variable | Description |
|---|---|
| `RCTL_CONFIG` | Config path (equivalent to `--config`) |
| `RCTL_VERBOSE` | `1` — enable verbose (equivalent to `--verbose`) |
| `RCTL_AUDIT_ENABLED` | `true`/`false` — toggle audit |
| `RCTL_DEFAULT_CLIENT` | Default client |

## What's next

- `rctl doctor` — check your config for errors
- `examples/config.yaml` — full example with all features
- `examples/simple-config.yaml` — minimal example
