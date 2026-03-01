package secrets

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"text/template"
)

// CommandProvider runs a command to fetch a secret.
type CommandProvider struct {
	// ProviderName is the full provider name, e.g., "command:op".
	ProviderName string
	// CmdTemplate is the command template string, e.g., "op read {{ .Path }}".
	CmdTemplate string
}

func (p *CommandProvider) Name() string { return p.ProviderName }

func (p *CommandProvider) Get(path string) (string, error) {
	// Render the command template
	t, err := template.New("cmd").Parse(p.CmdTemplate)
	if err != nil {
		return "", fmt.Errorf("parsing command template for %s: %w", p.ProviderName, err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, struct{ Path string }{Path: path}); err != nil {
		return "", fmt.Errorf("rendering command template for %s: %w", p.ProviderName, err)
	}

	cmdLine := buf.String()
	parts := strings.Fields(cmdLine)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command for provider %s", p.ProviderName)
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("command %q failed: %w (stderr: %s)", cmdLine, err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

// NewCommandProviders creates command providers from config.
func NewCommandProviders(commands map[string]string) []*CommandProvider {
	var providers []*CommandProvider
	for name, tmpl := range commands {
		providers = append(providers, &CommandProvider{
			ProviderName: "command:" + name,
			CmdTemplate:  tmpl,
		})
	}
	return providers
}
