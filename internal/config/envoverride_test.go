package config

import (
	"os"
	"testing"
)

func TestApplyEnvOverrides_AuditEnabled(t *testing.T) {
	root := &Root{Version: 1}

	os.Setenv("RCTL_AUDIT_ENABLED", "true")
	defer os.Unsetenv("RCTL_AUDIT_ENABLED")

	ApplyEnvOverrides(root)

	if root.Defaults.Policy == nil || root.Defaults.Policy.Audit == nil {
		t.Fatal("policy.audit should be created")
	}
	if !root.Defaults.Policy.Audit.Enabled {
		t.Error("audit should be enabled")
	}
}

func TestApplyEnvOverrides_AuditDisabled(t *testing.T) {
	root := &Root{
		Version: 1,
		Defaults: DomainConfig{
			Policy: &PolicyConfig{
				Audit: &AuditConfig{Enabled: true},
			},
		},
	}

	os.Setenv("RCTL_AUDIT_ENABLED", "false")
	defer os.Unsetenv("RCTL_AUDIT_ENABLED")

	ApplyEnvOverrides(root)

	if root.Defaults.Policy.Audit.Enabled {
		t.Error("audit should be disabled")
	}
}

func TestIsVerboseEnv(t *testing.T) {
	os.Unsetenv("RCTL_VERBOSE")
	if IsVerboseEnv() {
		t.Error("should be false when unset")
	}

	os.Setenv("RCTL_VERBOSE", "1")
	defer os.Unsetenv("RCTL_VERBOSE")
	if !IsVerboseEnv() {
		t.Error("should be true when set to 1")
	}
}

func TestGetDefaultClient(t *testing.T) {
	os.Unsetenv("RCTL_DEFAULT_CLIENT")
	if v := GetDefaultClient(); v != "" {
		t.Errorf("should be empty, got %q", v)
	}

	os.Setenv("RCTL_DEFAULT_CLIENT", "acme")
	defer os.Unsetenv("RCTL_DEFAULT_CLIENT")
	if v := GetDefaultClient(); v != "acme" {
		t.Errorf("should be acme, got %q", v)
	}
}
