# Shell Completion

rctl supports shell completion for bash, zsh, and fish.

## Setup

### Zsh

```bash
eval "$(rctl completion zsh)"
```

Add this line to `~/.zshrc` for persistent activation.

### Bash

```bash
eval "$(rctl completion bash)"
```

Add this line to `~/.bashrc` for persistent activation.

### Fish

```fish
rctl completion fish | source
```

For persistent activation:

```fish
rctl completion fish > ~/.config/fish/completions/rctl.fish
```

## What gets completed

### Clients

```
rctl <TAB>
→ lifedl  ldl  clients  domains  show  which  ...
```

Client names, their aliases, and rctl subcommands are suggested.

### Domains

```
rctl lifedl <TAB>
→ vpn  v  ansible  ans
```

Domain names for the selected client and their aliases are suggested.

### Commands

```
rctl lifedl vpn <TAB>
→ up  down  restart  status  logs  list  browse  vpnctl  browserctl
```

Suggested completions include:
- Domain profiles (`up`, `down`, `status`, ...)
- Global builtins
- Executables from the domain's `dir`
- Executables from `commands_path`

### Files (command arguments)

```
rctl lifedl ansible ansible-playbook <TAB>
→ inventory/  contrib/  run-profiles/  deploy.yml  site.yml

rctl lifedl ansible ansible-playbook inventory/<TAB>
→ inventory/arsmedica39/  inventory/dnkom/  inventory/medswiss/
```

Files are suggested from the domain's directory (`dir`), not the current directory.

### Subcommands

Completion also works for subcommands:

```
rctl show <TAB>        → clients
rctl show lifedl <TAB> → domains
rctl which lifedl vpn <TAB> → commands
```
