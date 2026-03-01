package resolve

import (
	"testing"

	"github.com/lavr/rctl/internal/config"
)

func TestFindClient_Found(t *testing.T) {
	cfg := &config.Root{
		Clients: map[string]*config.Client{
			"acme": {
				Domains: map[string]config.Domain{
					"vpn": {},
				},
			},
		},
	}

	c, err := FindClient(cfg, "acme")
	if err != nil {
		t.Fatalf("FindClient failed: %v", err)
	}
	if c == nil {
		t.Fatal("client is nil")
	}
	if _, ok := c.Domains["vpn"]; !ok {
		t.Error("domain 'vpn' not found in client")
	}
}

func TestFindClient_NotFound(t *testing.T) {
	cfg := &config.Root{
		Clients: map[string]*config.Client{
			"acme": {},
		},
	}

	_, err := FindClient(cfg, "nonexistent")
	if err == nil {
		t.Error("expected error for missing client")
	}
}

func TestFindDomain_Found(t *testing.T) {
	client := &config.Client{
		Domains: map[string]config.Domain{
			"vpn": {
				DomainConfig: config.DomainConfig{
					Dir: "/some/dir",
				},
			},
		},
	}

	d, err := FindDomain(client, "vpn")
	if err != nil {
		t.Fatalf("FindDomain failed: %v", err)
	}
	if d.Dir != "/some/dir" {
		t.Errorf("domain dir = %q, want %q", d.Dir, "/some/dir")
	}
}

func TestFindDomain_NotFound(t *testing.T) {
	client := &config.Client{
		Domains: map[string]config.Domain{
			"vpn": {},
		},
	}

	_, err := FindDomain(client, "nonexistent")
	if err == nil {
		t.Error("expected error for missing domain")
	}
}
