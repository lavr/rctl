package config

import (
	"fmt"
	"strings"
)

// ValidationError collects multiple validation issues.
type ValidationError struct {
	Errors []string
}

func (e *ValidationError) Error() string {
	return strings.Join(e.Errors, "; ")
}

func (e *ValidationError) add(msg string) {
	e.Errors = append(e.Errors, msg)
}

func (e *ValidationError) hasErrors() bool {
	return len(e.Errors) > 0
}

// Validate checks the config for common issues:
// - Alias uniqueness across clients
// - Alias uniqueness across domains within each client
// - Extends cycle detection
// - Version check
func Validate(root *Root) error {
	ve := &ValidationError{}

	if root.Version != 1 {
		ve.add(fmt.Sprintf("unsupported version: %d", root.Version))
	}

	validateAliasUniqueness(root, ve)
	validateExtendsCycles(root, ve)

	if ve.hasErrors() {
		return ve
	}
	return nil
}

func validateAliasUniqueness(root *Root, ve *ValidationError) {
	// Client alias uniqueness: no two clients should have the same alias,
	// and no alias should conflict with a client name.
	clientNames := make(map[string]string) // name/alias -> owning client
	for name := range root.Clients {
		if existing, ok := clientNames[name]; ok {
			ve.add(fmt.Sprintf("client name %q conflicts with %s", name, existing))
		}
		clientNames[name] = fmt.Sprintf("client %q", name)
	}

	for name, client := range root.Clients {
		for _, alias := range client.Aliases {
			if existing, ok := clientNames[alias]; ok {
				ve.add(fmt.Sprintf("client alias %q (from %q) conflicts with %s", alias, name, existing))
			} else {
				clientNames[alias] = fmt.Sprintf("client %q alias", name)
			}
		}

		// Domain alias uniqueness within each client
		domainNames := make(map[string]string)
		for dname := range client.Domains {
			domainNames[dname] = fmt.Sprintf("domain %q", dname)
		}
		for dname, domain := range client.Domains {
			for _, alias := range domain.Aliases {
				if existing, ok := domainNames[alias]; ok {
					ve.add(fmt.Sprintf("domain alias %q (from %q in client %q) conflicts with %s", alias, dname, name, existing))
				} else {
					domainNames[alias] = fmt.Sprintf("domain %q alias", dname)
				}
			}
		}
	}
}

func validateExtendsCycles(root *Root, ve *ValidationError) {
	for clientName, client := range root.Clients {
		for domainName, domain := range client.Domains {
			if domain.Extends == "" {
				continue
			}
			// Walk the extends chain, detect cycles
			visited := map[string]bool{domainName: true}
			current := domain.Extends
			for current != "" {
				if visited[current] {
					ve.add(fmt.Sprintf("extends cycle detected: domain %q in client %q", domainName, clientName))
					break
				}
				visited[current] = true
				next, ok := client.Domains[current]
				if !ok {
					ve.add(fmt.Sprintf("domain %q in client %q extends unknown domain %q", domainName, clientName, current))
					break
				}
				current = next.Extends
			}
		}
	}
}
