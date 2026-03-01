package resolve

import (
	"fmt"

	"github.com/lavr/rctl/internal/config"
)

// ResolveDomainExtends resolves the extends chain for a domain, merging
// parent configs into the domain. Returns the fully-resolved DomainConfig.
// Detects cycles.
func ResolveDomainExtends(client *config.Client, domainName string) (config.DomainConfig, error) {
	// Build the chain from most-base to most-derived
	chain, err := buildExtendsChain(client, domainName)
	if err != nil {
		return config.DomainConfig{}, err
	}

	// Merge chain: each layer overlays onto the previous
	var result config.DomainConfig
	for _, name := range chain {
		d := client.Domains[name]
		result = mergeDomainConfig(result, d.DomainConfig)
	}

	return result, nil
}

// buildExtendsChain returns the ordered list of domain names from root ancestor to the given domain.
func buildExtendsChain(client *config.Client, domainName string) ([]string, error) {
	var chain []string
	visited := make(map[string]bool)
	current := domainName

	for current != "" {
		if visited[current] {
			return nil, fmt.Errorf("extends cycle detected at domain %q", current)
		}
		visited[current] = true

		domain, ok := client.Domains[current]
		if !ok {
			return nil, fmt.Errorf("domain %q not found (referenced via extends)", current)
		}

		chain = append([]string{current}, chain...) // prepend: base first
		current = domain.Extends
	}

	return chain, nil
}

// mergeDomainConfig merges overlay onto base using the same rules as ComputeEffective.
func mergeDomainConfig(base, overlay config.DomainConfig) config.DomainConfig {
	result := base

	if overlay.Dir != "" {
		result.Dir = overlay.Dir
	}

	// Merge maps
	result.Env = mergeMaps(result.Env, overlay.Env)
	result.Profiles = mergeProfiles(result.Profiles, overlay.Profiles)
	result.Vars = mergeMaps(result.Vars, overlay.Vars)

	// Concat lists
	result.DefaultArgs = append(result.DefaultArgs, overlay.DefaultArgs...)
	result.PathAdd = append(result.PathAdd, overlay.PathAdd...)
	result.CommandsPath = append(result.CommandsPath, overlay.CommandsPath...)

	return result
}

func mergeMaps(base, overlay map[string]string) map[string]string {
	if len(base) == 0 && len(overlay) == 0 {
		return nil
	}
	result := make(map[string]string)
	for k, v := range base {
		result[k] = v
	}
	for k, v := range overlay {
		result[k] = v
	}
	return result
}

func mergeProfiles(base, overlay map[string]config.Profile) map[string]config.Profile {
	if len(base) == 0 && len(overlay) == 0 {
		return nil
	}
	result := make(map[string]config.Profile)
	for k, v := range base {
		result[k] = v
	}
	for k, v := range overlay {
		result[k] = v
	}
	return result
}
