package resolve

import (
	"strings"
	"testing"

	"github.com/lavr/rctl/internal/config"
)

func TestResolveDomainExtends_Simple(t *testing.T) {
	client := &config.Client{
		Domains: map[string]config.Domain{
			"vpn": {
				DomainConfig: config.DomainConfig{
					Dir: "/vpn",
					Env: map[string]string{"VPN": "1"},
					Profiles: map[string]config.Profile{
						"up": {Cmd: "wg-up"},
					},
				},
			},
			"vpn-prod": {
				DomainConfig: config.DomainConfig{
					Env: map[string]string{"ENV": "prod"},
				},
				Extends: "vpn",
			},
		},
	}

	result, err := ResolveDomainExtends(client, "vpn-prod")
	if err != nil {
		t.Fatalf("ResolveDomainExtends: %v", err)
	}

	if result.Dir != "/vpn" {
		t.Errorf("dir = %q, want /vpn", result.Dir)
	}
	if result.Env["VPN"] != "1" {
		t.Errorf("env VPN = %q, want 1", result.Env["VPN"])
	}
	if result.Env["ENV"] != "prod" {
		t.Errorf("env ENV = %q, want prod", result.Env["ENV"])
	}
	if result.Profiles["up"].Cmd != "wg-up" {
		t.Errorf("profile up cmd = %q, want wg-up", result.Profiles["up"].Cmd)
	}
}

func TestResolveDomainExtends_Chained(t *testing.T) {
	client := &config.Client{
		Domains: map[string]config.Domain{
			"base": {
				DomainConfig: config.DomainConfig{
					Dir: "/base",
					Env: map[string]string{"A": "1"},
				},
			},
			"middle": {
				DomainConfig: config.DomainConfig{
					Env: map[string]string{"B": "2"},
				},
				Extends: "base",
			},
			"leaf": {
				DomainConfig: config.DomainConfig{
					Env: map[string]string{"C": "3"},
				},
				Extends: "middle",
			},
		},
	}

	result, err := ResolveDomainExtends(client, "leaf")
	if err != nil {
		t.Fatalf("ResolveDomainExtends: %v", err)
	}

	if result.Dir != "/base" {
		t.Errorf("dir = %q, want /base", result.Dir)
	}
	if result.Env["A"] != "1" || result.Env["B"] != "2" || result.Env["C"] != "3" {
		t.Errorf("env = %v, want A=1 B=2 C=3", result.Env)
	}
}

func TestResolveDomainExtends_Cycle(t *testing.T) {
	client := &config.Client{
		Domains: map[string]config.Domain{
			"a": {Extends: "b"},
			"b": {Extends: "a"},
		},
	}

	_, err := ResolveDomainExtends(client, "a")
	if err == nil {
		t.Fatal("expected cycle error")
	}
	if !strings.Contains(err.Error(), "cycle") {
		t.Errorf("error = %q, should contain 'cycle'", err.Error())
	}
}

func TestResolveDomainExtends_NotFound(t *testing.T) {
	client := &config.Client{
		Domains: map[string]config.Domain{
			"vpn-prod": {Extends: "nonexistent"},
		},
	}

	_, err := ResolveDomainExtends(client, "vpn-prod")
	if err == nil {
		t.Fatal("expected not found error")
	}
}

func TestResolveDomainExtends_NoExtends(t *testing.T) {
	client := &config.Client{
		Domains: map[string]config.Domain{
			"vpn": {
				DomainConfig: config.DomainConfig{
					Dir: "/vpn",
				},
			},
		},
	}

	result, err := ResolveDomainExtends(client, "vpn")
	if err != nil {
		t.Fatalf("ResolveDomainExtends: %v", err)
	}
	if result.Dir != "/vpn" {
		t.Errorf("dir = %q, want /vpn", result.Dir)
	}
}

func TestResolveDomainExtends_ProfileInheritance(t *testing.T) {
	client := &config.Client{
		Domains: map[string]config.Domain{
			"base": {
				DomainConfig: config.DomainConfig{
					Profiles: map[string]config.Profile{
						"up":   {Cmd: "base-up"},
						"down": {Cmd: "base-down"},
					},
				},
			},
			"child": {
				DomainConfig: config.DomainConfig{
					Profiles: map[string]config.Profile{
						"up": {Cmd: "child-up", Args: []string{"--fast"}},
					},
				},
				Extends: "base",
			},
		},
	}

	result, err := ResolveDomainExtends(client, "child")
	if err != nil {
		t.Fatalf("ResolveDomainExtends: %v", err)
	}

	if result.Profiles["up"].Cmd != "child-up" {
		t.Errorf("up cmd = %q, want child-up", result.Profiles["up"].Cmd)
	}
	if result.Profiles["down"].Cmd != "base-down" {
		t.Errorf("down cmd = %q, want base-down", result.Profiles["down"].Cmd)
	}
}
