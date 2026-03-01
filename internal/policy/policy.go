package policy

import (
	"fmt"
	"os"
	"strings"

	"github.com/lavr/rctl/internal/config"
)

// Denial describes why a command was denied.
type Denial struct {
	Reason string
}

func (d *Denial) Error() string {
	return d.Reason
}

// Check evaluates policy rules against the given command.
// Returns nil if allowed, or a Denial explaining why it's denied.
func Check(pol *config.PolicyConfig, cmd string, execPath string) *Denial {
	if pol == nil {
		return nil
	}

	// Check deny_commands
	for _, denied := range pol.DenyCommands {
		if cmd == denied {
			return &Denial{Reason: fmt.Sprintf("command %q is denied by policy", cmd)}
		}
	}

	// Check allow_commands (strict mode: if non-empty, only listed commands are allowed)
	if len(pol.AllowCommands) > 0 {
		allowed := false
		for _, a := range pol.AllowCommands {
			if cmd == a {
				allowed = true
				break
			}
		}
		if !allowed {
			return &Denial{Reason: fmt.Sprintf("command %q is not in allow_commands list", cmd)}
		}
	}

	// Check allow_path_prefixes
	if len(pol.AllowPathPrefixes) > 0 {
		allowed := false
		for _, prefix := range pol.AllowPathPrefixes {
			if strings.HasPrefix(execPath, prefix) {
				allowed = true
				break
			}
		}
		if !allowed {
			return &Denial{Reason: fmt.Sprintf("executable path %q does not match any allowed path prefix", execPath)}
		}
	}

	// Check require_tty
	if pol.RequireTTY {
		if !isTerminal() {
			return &Denial{Reason: "policy requires a TTY but stdin is not a terminal"}
		}
	}

	return nil
}

func isTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
