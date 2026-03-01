# Configuration

## Default path

`~/.config/rctl/config.yaml`

Defined in `internal/config/load.go`. Tilde is expanded to the user's home directory.

## Specifying the config path

Path resolution priority (highest to lowest):

1. **CLI flag** `--config <path>` — highest priority
2. **Environment variable** `RCTL_CONFIG` — medium
3. **Default path** `~/.config/rctl/config.yaml` — fallback

Logic in `internal/cli/app.go`.

## Multiple configs and merging

Two mechanisms are supported:

### 1. `includes` — explicit glob patterns

```yaml
version: 1
includes:
  - "~/.config/rctl/clients/*.yaml"
```

All files matching the pattern are loaded and merged into the base config.

### 2. `config.d/` — automatic directory

`.yaml`/`.yml` files from the `config.d/` directory located next to the main config (i.e. `~/.config/rctl/config.d/`) are automatically picked up and merged. Files are sorted by name.

## Loading pipeline

```
1. Load base config as yaml.Node
2. Extract and expand includes
3. Load files from config.d/
4. Recursive merge: base → includes → config.d
5. Decode into Root struct
6. Validate version and expand ~ in paths
```

## Merge strategy

- **Maps:** overlay keys overwrite base keys
- **Lists:** concatenation by default (`[a, b]` + `[c]` = `[a, b, c]`)
- **Lists with `@override`:** full replacement (`default_args@override: [c]` = `[c]`)
- **Scalars:** overlay replaces base

Merge logic in `internal/config/merge.go`.

## Additional environment variables

After loading the config, additional overrides from environment variables are applied (`internal/config/envoverride.go`):

| Variable | Description |
|----------|-------------|
| `RCTL_CONFIG` | Config path (resolved at CLI level) |
| `RCTL_AUDIT_ENABLED` | Enable/disable audit (`true`/`false`) |
| `RCTL_VERBOSE` | Verbose logging |
| `RCTL_DEFAULT_CLIENT` | Default client |
