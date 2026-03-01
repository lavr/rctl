package config

import (
	"strings"
	"testing"
	"text/template"
)

func TestRenderEffective_Vars(t *testing.T) {
	eff := &EffectiveConfig{
		Dir: "/work/{{ .Client.Name }}/{{ .Domain.Name }}",
		Env: map[string]string{
			"REGION": "{{ .Vars.region }}",
		},
		Vars: map[string]string{"region": "eu"},
	}

	ctx := &TemplateContext{}
	ctx.Client.Name = "acme"
	ctx.Domain.Name = "vpn"
	ctx.Vars = eff.Vars
	ctx.Env = map[string]string{}

	err := RenderEffective(eff, ctx, template.FuncMap{})
	if err != nil {
		t.Fatalf("RenderEffective: %v", err)
	}

	if eff.Dir != "/work/acme/vpn" {
		t.Errorf("dir = %q, want /work/acme/vpn", eff.Dir)
	}
	if eff.Env["REGION"] != "eu" {
		t.Errorf("env REGION = %q, want eu", eff.Env["REGION"])
	}
}

func TestRenderEffective_Host(t *testing.T) {
	eff := &EffectiveConfig{
		Env: map[string]string{
			"OS": "{{ .Host.OS }}",
		},
		Vars: map[string]string{},
	}

	ctx := &TemplateContext{}
	ctx.Host.OS = "linux"
	ctx.Vars = map[string]string{}
	ctx.Env = map[string]string{}

	err := RenderEffective(eff, ctx, template.FuncMap{})
	if err != nil {
		t.Fatalf("RenderEffective: %v", err)
	}

	if eff.Env["OS"] != "linux" {
		t.Errorf("env OS = %q, want linux", eff.Env["OS"])
	}
}

func TestRenderEffective_NoTemplate(t *testing.T) {
	eff := &EffectiveConfig{
		Dir: "/simple/path",
		Env: map[string]string{"KEY": "value"},
		Vars: map[string]string{},
	}

	ctx := &TemplateContext{}
	ctx.Vars = map[string]string{}
	ctx.Env = map[string]string{}

	err := RenderEffective(eff, ctx, template.FuncMap{})
	if err != nil {
		t.Fatalf("RenderEffective: %v", err)
	}

	if eff.Dir != "/simple/path" {
		t.Errorf("dir = %q, want /simple/path", eff.Dir)
	}
}

func TestRenderEffective_ErrorContainsFieldPath(t *testing.T) {
	eff := &EffectiveConfig{
		Env: map[string]string{
			"BAD": "{{ .NonExistent.Field }}",
		},
		Vars: map[string]string{},
	}

	ctx := &TemplateContext{}
	ctx.Vars = map[string]string{}
	ctx.Env = map[string]string{}

	err := RenderEffective(eff, ctx, template.FuncMap{})
	if err == nil {
		t.Fatal("expected error for bad template")
	}
	if !strings.Contains(err.Error(), "env.BAD") {
		t.Errorf("error = %q, should contain field path 'env.BAD'", err.Error())
	}
}

func TestRenderEffective_ProfileArgs(t *testing.T) {
	eff := &EffectiveConfig{
		Profiles: map[string]Profile{
			"up": {
				Cmd:  "vpn-{{ .Vars.type }}",
				Args: []string{"--region", "{{ .Vars.region }}"},
			},
		},
		Vars: map[string]string{"type": "wg", "region": "eu"},
	}

	ctx := &TemplateContext{}
	ctx.Vars = eff.Vars
	ctx.Env = map[string]string{}

	err := RenderEffective(eff, ctx, template.FuncMap{})
	if err != nil {
		t.Fatalf("RenderEffective: %v", err)
	}

	p := eff.Profiles["up"]
	if p.Cmd != "vpn-wg" {
		t.Errorf("profile cmd = %q, want vpn-wg", p.Cmd)
	}
	if p.Args[1] != "eu" {
		t.Errorf("profile args[1] = %q, want eu", p.Args[1])
	}
}

func TestRenderEffective_DefaultArgs(t *testing.T) {
	eff := &EffectiveConfig{
		DefaultArgs: []string{"--client={{ .Client.Name }}"},
		Vars:        map[string]string{},
	}

	ctx := &TemplateContext{}
	ctx.Client.Name = "acme"
	ctx.Vars = map[string]string{}
	ctx.Env = map[string]string{}

	err := RenderEffective(eff, ctx, template.FuncMap{})
	if err != nil {
		t.Fatalf("RenderEffective: %v", err)
	}

	if eff.DefaultArgs[0] != "--client=acme" {
		t.Errorf("default_args[0] = %q, want --client=acme", eff.DefaultArgs[0])
	}
}
