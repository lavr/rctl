package cli

import (
	"io"
	"log"
	"os"

	"github.com/lavr/rctl/internal/config"
)

// App holds shared state for the CLI.
type App struct {
	Version string
	Cfg     *config.Root
	Verbose *log.Logger

	// Flags
	ConfigPath   string
	VerboseFlag  bool
	DryRun       bool
	PrintCmdline bool
	PrintEnv     bool

	Out io.Writer
	Err io.Writer
}

// NewApp creates a new App with defaults.
func NewApp(version string) *App {
	return &App{
		Version: version,
		Verbose: log.New(io.Discard, "", 0),
		Out:     os.Stdout,
		Err:     os.Stderr,
	}
}

// LoadConfig loads the config, applies env overrides, and validates.
func (a *App) LoadConfig() error {
	path := a.ConfigPath
	if path == "" {
		if envPath := os.Getenv("RCTL_CONFIG"); envPath != "" {
			path = envPath
		} else {
			path = config.DefaultConfigPath()
		}
	}
	cfg, err := config.Load(path)
	if err != nil {
		return err
	}

	// Apply env overrides
	config.ApplyEnvOverrides(cfg)

	// Validate
	if err := config.Validate(cfg); err != nil {
		return err
	}

	a.Cfg = cfg
	return nil
}

// SetupVerbose configures the verbose logger based on the flag or env.
func (a *App) SetupVerbose() {
	if a.VerboseFlag || config.IsVerboseEnv() {
		a.Verbose = log.New(a.Err, "[rctl] ", 0)
	}
}
