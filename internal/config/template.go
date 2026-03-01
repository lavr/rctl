package config

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"
	"text/template"
	"time"
)

// TemplateContext provides data available in templates.
type TemplateContext struct {
	Client struct {
		Name string
		Tags []string
	}
	Domain struct {
		Name string
	}
	Dir  string
	Vars map[string]string
	Env  map[string]string
	Host struct {
		Name string
		OS   string
		Arch string
	}
	Time struct {
		Now  string
		Date string
	}
}

// NewTemplateContext creates a TemplateContext populated with runtime data.
func NewTemplateContext(clientName string, clientTags []string, domainName string, eff *EffectiveConfig) *TemplateContext {
	ctx := &TemplateContext{}
	ctx.Client.Name = clientName
	ctx.Client.Tags = clientTags
	ctx.Domain.Name = domainName
	ctx.Dir = eff.Dir
	ctx.Vars = eff.Vars

	// Build env from current process + effective overrides
	ctx.Env = make(map[string]string)
	for _, e := range os.Environ() {
		if k, v, ok := strings.Cut(e, "="); ok {
			ctx.Env[k] = v
		}
	}
	for k, v := range eff.Env {
		ctx.Env[k] = v
	}

	hostname, _ := os.Hostname()
	ctx.Host.Name = hostname
	ctx.Host.OS = runtime.GOOS
	ctx.Host.Arch = runtime.GOARCH

	now := time.Now()
	ctx.Time.Now = now.Format(time.RFC3339)
	ctx.Time.Date = now.Format("2006-01-02")

	return ctx
}

// RenderEffective renders all template strings in the effective config.
func RenderEffective(eff *EffectiveConfig, ctx *TemplateContext, funcMap template.FuncMap) error {
	var err error

	// Render dir
	if eff.Dir, err = renderField("dir", eff.Dir, ctx, funcMap); err != nil {
		return err
	}

	// Render env values
	for k, v := range eff.Env {
		if eff.Env[k], err = renderField(fmt.Sprintf("env.%s", k), v, ctx, funcMap); err != nil {
			return err
		}
	}

	// Render default_args
	for i, v := range eff.DefaultArgs {
		if eff.DefaultArgs[i], err = renderField(fmt.Sprintf("default_args[%d]", i), v, ctx, funcMap); err != nil {
			return err
		}
	}

	// Render commands_path
	for i, v := range eff.CommandsPath {
		if eff.CommandsPath[i], err = renderField(fmt.Sprintf("commands_path[%d]", i), v, ctx, funcMap); err != nil {
			return err
		}
	}

	// Render path_add
	for i, v := range eff.PathAdd {
		if eff.PathAdd[i], err = renderField(fmt.Sprintf("path_add[%d]", i), v, ctx, funcMap); err != nil {
			return err
		}
	}

	// Render profile args
	for name, p := range eff.Profiles {
		if p.Cmd, err = renderField(fmt.Sprintf("profiles.%s.cmd", name), p.Cmd, ctx, funcMap); err != nil {
			return err
		}
		for i, v := range p.Args {
			if p.Args[i], err = renderField(fmt.Sprintf("profiles.%s.args[%d]", name, i), v, ctx, funcMap); err != nil {
				return err
			}
		}
		eff.Profiles[name] = p
	}

	// Expand ~ in rendered paths (templates may produce paths with ~ from vars)
	expandRenderedPaths(eff)

	return nil
}

// expandRenderedPaths expands ~ in path fields after template rendering.
func expandRenderedPaths(eff *EffectiveConfig) {
	eff.Dir = ExpandTilde(eff.Dir)
	for i, p := range eff.PathAdd {
		eff.PathAdd[i] = ExpandTilde(p)
	}
	for i, p := range eff.CommandsPath {
		eff.CommandsPath[i] = ExpandTilde(p)
	}
}

func renderField(fieldPath, tmplStr string, ctx *TemplateContext, funcMap template.FuncMap) (string, error) {
	if !strings.Contains(tmplStr, "{{") {
		return tmplStr, nil
	}

	t, err := template.New(fieldPath).Funcs(funcMap).Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("template error in %s: %w", fieldPath, err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, ctx); err != nil {
		return "", fmt.Errorf("template render error in %s: %w", fieldPath, err)
	}

	return buf.String(), nil
}

// DefaultFuncMap returns the default template function map.
func DefaultFuncMap() template.FuncMap {
	return template.FuncMap{
		"env": os.Getenv,
	}
}
