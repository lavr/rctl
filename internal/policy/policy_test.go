package policy

import (
	"testing"

	"github.com/lavr/rctl/internal/config"
)

func TestCheck_NilPolicy(t *testing.T) {
	denial := Check(nil, "echo", "/bin/echo")
	if denial != nil {
		t.Errorf("expected nil denial for nil policy, got: %s", denial.Reason)
	}
}

func TestCheck_DenyCommands(t *testing.T) {
	pol := &config.PolicyConfig{
		DenyCommands: []string{"rm", "dd"},
	}

	if d := Check(pol, "rm", "/bin/rm"); d == nil {
		t.Error("expected denial for denied command 'rm'")
	}
	if d := Check(pol, "echo", "/bin/echo"); d != nil {
		t.Errorf("unexpected denial for 'echo': %s", d.Reason)
	}
}

func TestCheck_AllowCommandsStrict(t *testing.T) {
	pol := &config.PolicyConfig{
		AllowCommands: []string{"echo", "ls"},
	}

	if d := Check(pol, "echo", "/bin/echo"); d != nil {
		t.Errorf("unexpected denial for allowed 'echo': %s", d.Reason)
	}
	if d := Check(pol, "rm", "/bin/rm"); d == nil {
		t.Error("expected denial for non-allowed 'rm'")
	}
}

func TestCheck_AllowPathPrefixes(t *testing.T) {
	pol := &config.PolicyConfig{
		AllowPathPrefixes: []string{"/usr/bin/", "/usr/local/bin/"},
	}

	if d := Check(pol, "echo", "/usr/bin/echo"); d != nil {
		t.Errorf("unexpected denial: %s", d.Reason)
	}
	if d := Check(pol, "evil", "/tmp/evil"); d == nil {
		t.Error("expected denial for path outside allowed prefixes")
	}
}

func TestCheck_Combined(t *testing.T) {
	pol := &config.PolicyConfig{
		AllowCommands:     []string{"echo", "ls"},
		DenyCommands:      []string{"echo"},
		AllowPathPrefixes: []string{"/usr/bin/"},
	}

	// echo is in allow AND deny — deny takes priority (checked first)
	if d := Check(pol, "echo", "/usr/bin/echo"); d == nil {
		t.Error("expected denial: deny should take priority")
	}

	// ls is allowed and path matches
	if d := Check(pol, "ls", "/usr/bin/ls"); d != nil {
		t.Errorf("unexpected denial for ls: %s", d.Reason)
	}

	// ls with wrong path
	if d := Check(pol, "ls", "/tmp/ls"); d == nil {
		t.Error("expected denial: path doesn't match")
	}
}
