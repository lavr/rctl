package resolve

import (
	"strings"
	"testing"

	"github.com/lavr/rctl/internal/config"
)

func TestFindClientByNameOrAlias_ExactName(t *testing.T) {
	cfg := &config.Root{
		Clients: map[string]*config.Client{
			"acme": {Aliases: []string{"ac"}},
		},
	}

	name, client, err := FindClientByNameOrAlias(cfg, "acme")
	if err != nil {
		t.Fatalf("FindClientByNameOrAlias: %v", err)
	}
	if name != "acme" {
		t.Errorf("name = %q, want acme", name)
	}
	if client == nil {
		t.Fatal("client is nil")
	}
}

func TestFindClientByNameOrAlias_Alias(t *testing.T) {
	cfg := &config.Root{
		Clients: map[string]*config.Client{
			"acme": {Aliases: []string{"ac", "a"}},
		},
	}

	name, _, err := FindClientByNameOrAlias(cfg, "ac")
	if err != nil {
		t.Fatalf("FindClientByNameOrAlias: %v", err)
	}
	if name != "acme" {
		t.Errorf("name = %q, want acme", name)
	}
}

func TestFindClientByNameOrAlias_NotFound(t *testing.T) {
	cfg := &config.Root{
		Clients: map[string]*config.Client{
			"acme": {},
		},
	}

	_, _, err := FindClientByNameOrAlias(cfg, "xyz")
	if err == nil {
		t.Error("expected error for missing client")
	}
}

func TestFindClientByNameOrAlias_Ambiguous(t *testing.T) {
	cfg := &config.Root{
		Clients: map[string]*config.Client{
			"acme":  {Aliases: []string{"shared"}},
			"globex": {Aliases: []string{"shared"}},
		},
	}

	_, _, err := FindClientByNameOrAlias(cfg, "shared")
	if err == nil {
		t.Fatal("expected error for ambiguous alias")
	}
	if !strings.Contains(err.Error(), "ambiguous") {
		t.Errorf("error = %q, should contain 'ambiguous'", err.Error())
	}
}

func TestFindDomainByNameOrAlias_ExactName(t *testing.T) {
	client := &config.Client{
		Domains: map[string]config.Domain{
			"vpn": {DomainConfig: config.DomainConfig{Dir: "/vpn"}},
		},
	}

	name, domain, err := FindDomainByNameOrAlias(client, "vpn")
	if err != nil {
		t.Fatalf("FindDomainByNameOrAlias: %v", err)
	}
	if name != "vpn" {
		t.Errorf("name = %q, want vpn", name)
	}
	if domain.Dir != "/vpn" {
		t.Errorf("dir = %q, want /vpn", domain.Dir)
	}
}

func TestFindDomainByNameOrAlias_Alias(t *testing.T) {
	client := &config.Client{
		Domains: map[string]config.Domain{
			"vpn": {
				DomainConfig: config.DomainConfig{Dir: "/vpn"},
				Aliases:      []string{"v"},
			},
		},
	}

	name, _, err := FindDomainByNameOrAlias(client, "v")
	if err != nil {
		t.Fatalf("FindDomainByNameOrAlias: %v", err)
	}
	if name != "vpn" {
		t.Errorf("name = %q, want vpn", name)
	}
}

func TestFindDomainByNameOrAlias_NotFound(t *testing.T) {
	client := &config.Client{
		Domains: map[string]config.Domain{
			"vpn": {},
		},
	}

	_, _, err := FindDomainByNameOrAlias(client, "xyz")
	if err == nil {
		t.Error("expected error")
	}
}
