package config

// Root is the top-level configuration structure (version: 1).
type Root struct {
	Version  int                `yaml:"version"`
	Includes []string           `yaml:"includes,omitempty"`
	Defaults DomainConfig       `yaml:"defaults,omitempty"`
	Vars     map[string]string  `yaml:"vars,omitempty"`
	Clients  map[string]*Client `yaml:"clients,omitempty"`
	Builtins map[string]Builtin `yaml:"builtins,omitempty"`
	Secrets  SecretsConfig      `yaml:"secrets,omitempty"`
}

// Client represents a single client entry.
type Client struct {
	Aliases  []string          `yaml:"aliases,omitempty"`
	Tags     []string          `yaml:"tags,omitempty"`
	Defaults DomainConfig      `yaml:"defaults,omitempty"`
	Domains  map[string]Domain `yaml:"domains,omitempty"`
}

// Domain represents a single domain entry within a client.
type Domain struct {
	DomainConfig `yaml:",inline"`
	Aliases      []string `yaml:"aliases,omitempty"`
	Extends      string   `yaml:"extends,omitempty"`
}

// DomainConfig holds the common configuration fields shared across
// defaults, client.defaults, and domain levels.
type DomainConfig struct {
	Dir          string             `yaml:"dir,omitempty"`
	Env          map[string]string  `yaml:"env,omitempty"`
	DefaultArgs  []string           `yaml:"default_args,omitempty"`
	PathAdd      []string           `yaml:"path_add,omitempty"`
	CommandsPath []string           `yaml:"commands_path,omitempty"`
	Profiles     map[string]Profile `yaml:"profiles,omitempty"`
	Vars         map[string]string  `yaml:"vars,omitempty"`
	TaskRunner   string             `yaml:"task_runner,omitempty"`
	Policy       *PolicyConfig      `yaml:"policy,omitempty"`
}

// Profile defines a named command shortcut.
type Profile struct {
	Cmd  string   `yaml:"cmd" json:"cmd"`
	Args []string `yaml:"args,omitempty" json:"args,omitempty"`
}

// Builtin defines a global command alias.
type Builtin struct {
	Cmd  string   `yaml:"cmd"`
	Args []string `yaml:"args,omitempty"`
}

// PolicyConfig holds allow/deny rules.
type PolicyConfig struct {
	AllowPathPrefixes []string     `yaml:"allow_path_prefixes,omitempty"`
	DenyCommands      []string     `yaml:"deny_commands,omitempty"`
	AllowCommands     []string     `yaml:"allow_commands,omitempty"`
	RequireTTY        bool         `yaml:"require_tty,omitempty"`
	Audit             *AuditConfig `yaml:"audit,omitempty"`
}

// AuditConfig controls audit logging.
type AuditConfig struct {
	Enabled       bool     `yaml:"enabled,omitempty"`
	Sink          string   `yaml:"sink,omitempty"`
	File          string   `yaml:"file,omitempty"`
	RedactEnvKeys []string `yaml:"redact_env_keys,omitempty"`
}

// SecretsConfig holds secret provider settings.
type SecretsConfig struct {
	Providers SecretsProviders `yaml:"providers,omitempty"`
}

// SecretsProviders lists available secret providers.
type SecretsProviders struct {
	Env     *EnvSecretProvider     `yaml:"env,omitempty"`
	Command *CommandSecretProvider `yaml:"command,omitempty"`
}

// EnvSecretProvider reads secrets from environment variables.
type EnvSecretProvider struct {
	Enabled bool `yaml:"enabled,omitempty"`
}

// CommandSecretProvider runs commands to fetch secrets.
type CommandSecretProvider struct {
	Enabled  bool              `yaml:"enabled,omitempty"`
	Commands map[string]string `yaml:"commands,omitempty"`
}

// EffectiveConfig is the fully-merged config for a specific client+domain.
type EffectiveConfig struct {
	Dir          string
	Env          map[string]string
	DefaultArgs  []string
	PathAdd      []string
	CommandsPath []string
	Profiles     map[string]Profile
	Vars         map[string]string
	TaskRunner   string
}
