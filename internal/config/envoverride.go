package config

import (
	"os"
	"strings"
)

// ApplyEnvOverrides applies RCTL_* environment variable overrides to the config.
func ApplyEnvOverrides(root *Root) {
	// RCTL_CONFIG is handled at the CLI level (flag default), not here.

	// RCTL_AUDIT_ENABLED
	if v := os.Getenv("RCTL_AUDIT_ENABLED"); v != "" {
		enabled := strings.EqualFold(v, "true") || v == "1"
		if root.Defaults.Policy == nil {
			root.Defaults.Policy = &PolicyConfig{}
		}
		if root.Defaults.Policy.Audit == nil {
			root.Defaults.Policy.Audit = &AuditConfig{}
		}
		root.Defaults.Policy.Audit.Enabled = enabled
	}

	// RCTL_DEFAULT_CLIENT is read at dispatch time, not stored in config.
}

// GetDefaultClient returns the RCTL_DEFAULT_CLIENT env var, if set.
func GetDefaultClient() string {
	return os.Getenv("RCTL_DEFAULT_CLIENT")
}

// IsVerboseEnv returns true if RCTL_VERBOSE is set.
func IsVerboseEnv() bool {
	v := os.Getenv("RCTL_VERBOSE")
	return v == "1" || strings.EqualFold(v, "true")
}
