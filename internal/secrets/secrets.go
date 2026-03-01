package secrets

import (
	"fmt"
	"sync"
)

// Provider is the interface for secret backends.
type Provider interface {
	// Name returns the provider name (e.g., "env", "command:op").
	Name() string
	// Get retrieves a secret value given a path/key.
	Get(path string) (string, error)
}

// Registry manages secret providers and caches resolved secrets.
type Registry struct {
	providers map[string]Provider
	cache     map[string]string
	mu        sync.Mutex
}

// NewRegistry creates a new empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
		cache:     make(map[string]string),
	}
}

// Register adds a provider to the registry.
func (r *Registry) Register(p Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[p.Name()] = p
}

// Get retrieves a secret, using the cache if available.
// The key format is "provider:path", e.g., "env:MY_TOKEN" or "command:op:vaults/xxx".
func (r *Registry) Get(providerName, path string) (string, error) {
	cacheKey := providerName + ":" + path

	r.mu.Lock()
	if v, ok := r.cache[cacheKey]; ok {
		r.mu.Unlock()
		return v, nil
	}
	r.mu.Unlock()

	p, ok := r.providers[providerName]
	if !ok {
		return "", fmt.Errorf("unknown secret provider %q", providerName)
	}

	val, err := p.Get(path)
	if err != nil {
		return "", fmt.Errorf("secret %s:%s: %w", providerName, path, err)
	}

	r.mu.Lock()
	r.cache[cacheKey] = val
	r.mu.Unlock()

	return val, nil
}

// SecretFunc returns a template function for use in FuncMap.
// Usage in templates: {{ secret "env" "MY_TOKEN" }} or {{ secret "command:op" "path" }}
func (r *Registry) SecretFunc() func(string, ...string) (string, error) {
	return func(provider string, args ...string) (string, error) {
		if len(args) == 0 {
			// Single-arg form: "env:MY_TOKEN"
			return r.Get(provider, "")
		}
		return r.Get(provider, args[0])
	}
}
