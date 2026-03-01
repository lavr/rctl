package resolve

import (
	"fmt"

	"github.com/lavr/rctl/internal/config"
)

// FindClientByNameOrAlias looks up a client by name first, then by alias.
func FindClientByNameOrAlias(cfg *config.Root, nameOrAlias string) (string, *config.Client, error) {
	// Try exact name first
	if c, ok := cfg.Clients[nameOrAlias]; ok {
		return nameOrAlias, c, nil
	}

	// Search aliases
	var matches []string
	for name, client := range cfg.Clients {
		for _, alias := range client.Aliases {
			if alias == nameOrAlias {
				matches = append(matches, name)
			}
		}
	}

	switch len(matches) {
	case 0:
		return "", nil, fmt.Errorf("client %q not found", nameOrAlias)
	case 1:
		return matches[0], cfg.Clients[matches[0]], nil
	default:
		return "", nil, fmt.Errorf("ambiguous client alias %q: matches %v", nameOrAlias, matches)
	}
}

// FindDomainByNameOrAlias looks up a domain by name first, then by alias.
func FindDomainByNameOrAlias(client *config.Client, nameOrAlias string) (string, config.Domain, error) {
	// Try exact name first
	if d, ok := client.Domains[nameOrAlias]; ok {
		return nameOrAlias, d, nil
	}

	// Search aliases
	var matches []string
	for name, domain := range client.Domains {
		for _, alias := range domain.Aliases {
			if alias == nameOrAlias {
				matches = append(matches, name)
			}
		}
	}

	switch len(matches) {
	case 0:
		return "", config.Domain{}, fmt.Errorf("domain %q not found", nameOrAlias)
	case 1:
		return matches[0], client.Domains[matches[0]], nil
	default:
		return "", config.Domain{}, fmt.Errorf("ambiguous domain alias %q: matches %v", nameOrAlias, matches)
	}
}
