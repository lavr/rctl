package secrets

import (
	"fmt"
	"os"
)

// EnvProvider reads secrets from environment variables.
type EnvProvider struct{}

func (p *EnvProvider) Name() string { return "env" }

func (p *EnvProvider) Get(key string) (string, error) {
	val := os.Getenv(key)
	if val == "" {
		return "", fmt.Errorf("environment variable %q is not set or empty", key)
	}
	return val, nil
}
