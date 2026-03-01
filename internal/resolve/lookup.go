package resolve

import (
	"fmt"

	"github.com/lavr/rctl/internal/config"
)

// FindClient looks up a client by exact name.
func FindClient(cfg *config.Root, name string) (*config.Client, error) {
	c, ok := cfg.Clients[name]
	if !ok {
		return nil, fmt.Errorf("client %q not found", name)
	}
	return c, nil
}

// FindDomain looks up a domain by exact name within a client.
func FindDomain(client *config.Client, name string) (config.Domain, error) {
	d, ok := client.Domains[name]
	if !ok {
		return config.Domain{}, fmt.Errorf("domain %q not found", name)
	}
	return d, nil
}
