package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/lavr/rctl/internal/config"
	"github.com/lavr/rctl/internal/resolve"
	"github.com/spf13/cobra"
)

func doctorCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Run diagnostic checks on configuration",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			checks := []struct {
				name string
				fn   func(*App) (string, bool)
			}{
				{"Config file", checkConfigFile},
				{"Config version", checkConfigVersion},
				{"Includes", checkIncludes},
				{"Alias uniqueness", checkAliases},
				{"Extends cycles", checkExtendsCycles},
				{"Directories", checkDirectories},
				{"Commands paths", checkCommandsPaths},
				{"Task runners", checkTaskRunners},
				{"Templates", checkTemplates},
				{"Secret providers", checkProviders},
				{"Audit file", checkAuditFile},
			}

			passed := 0
			failed := 0

			for _, c := range checks {
				msg, ok := c.fn(app)
				if ok {
					fmt.Fprintf(app.Out, "  [ok] %s: %s\n", c.name, msg)
					passed++
				} else {
					fmt.Fprintf(app.Out, "  [!!] %s: %s\n", c.name, msg)
					failed++
				}
			}

			fmt.Fprintf(app.Out, "\n%d passed, %d failed\n", passed, failed)
			if failed > 0 {
				setExitCode(1)
			}
			return nil
		},
	}
}

func checkConfigFile(app *App) (string, bool) {
	path := app.ConfigPath
	if path == "" {
		path = config.DefaultConfigPath()
	}
	if _, err := os.Stat(path); err != nil {
		return fmt.Sprintf("%s not found", path), false
	}
	return path, true
}

func checkConfigVersion(app *App) (string, bool) {
	if app.Cfg == nil {
		return "no config loaded", false
	}
	if app.Cfg.Version != 1 {
		return fmt.Sprintf("unsupported version %d", app.Cfg.Version), false
	}
	return "version 1", true
}

func checkIncludes(app *App) (string, bool) {
	if app.Cfg == nil {
		return "no config loaded", false
	}
	if len(app.Cfg.Includes) == 0 {
		return "no includes configured", true
	}

	var files []string
	for _, pattern := range app.Cfg.Includes {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Sprintf("invalid glob %q: %v", pattern, err), false
		}
		files = append(files, matches...)
	}
	if len(files) == 0 {
		return "no files matched", true
	}
	return fmt.Sprintf("%d files: %v", len(files), files), true
}

func checkAliases(app *App) (string, bool) {
	if app.Cfg == nil {
		return "no config loaded", false
	}
	if err := config.Validate(app.Cfg); err != nil {
		return err.Error(), false
	}
	return "no conflicts", true
}

func checkExtendsCycles(app *App) (string, bool) {
	if app.Cfg == nil {
		return "no config loaded", false
	}

	extendsCount := 0
	for _, client := range app.Cfg.Clients {
		for _, domain := range client.Domains {
			if domain.Extends != "" {
				extendsCount++
			}
		}
	}

	// Validate already checks for cycles
	if err := config.Validate(app.Cfg); err != nil {
		return err.Error(), false
	}

	if extendsCount == 0 {
		return "no extends used", true
	}
	return fmt.Sprintf("%d extends relations, no cycles", extendsCount), true
}

func checkDirectories(app *App) (string, bool) {
	if app.Cfg == nil {
		return "no config loaded", false
	}

	var missing []string
	for clientName, client := range app.Cfg.Clients {
		for dname := range client.Domains {
			eff, err := renderEffectiveForDoctor(app.Cfg, clientName, client, dname)
			if err != nil {
				continue
			}
			if eff.Dir != "" {
				if _, err := os.Stat(eff.Dir); err != nil {
					missing = append(missing, fmt.Sprintf("%s: %s", dname, eff.Dir))
				}
			}
		}
	}

	if len(missing) > 0 {
		return fmt.Sprintf("%d missing: %v", len(missing), missing), false
	}
	return "all directories exist", true
}

func checkCommandsPaths(app *App) (string, bool) {
	if app.Cfg == nil {
		return "no config loaded", false
	}

	total := 0
	var missing []string
	for clientName, client := range app.Cfg.Clients {
		for dname := range client.Domains {
			eff, err := renderEffectiveForDoctor(app.Cfg, clientName, client, dname)
			if err != nil {
				continue
			}
			for _, p := range eff.CommandsPath {
				total++
				if _, err := os.Stat(p); err != nil {
					missing = append(missing, fmt.Sprintf("%s: %s", dname, p))
				}
			}
		}
	}

	if len(missing) > 0 {
		return fmt.Sprintf("%d missing: %v", len(missing), missing), false
	}
	if total == 0 {
		return "no commands_path configured", true
	}
	return fmt.Sprintf("%d paths, all exist", total), true
}

func checkTaskRunners(app *App) (string, bool) {
	if app.Cfg == nil {
		return "no config loaded", false
	}

	var configured []string
	var missing []string
	for clientName, client := range app.Cfg.Clients {
		for dname := range client.Domains {
			eff, err := renderEffectiveForDoctor(app.Cfg, clientName, client, dname)
			if err != nil {
				continue
			}
			if eff.TaskRunner == "" {
				continue
			}
			runner := eff.TaskRunner
			if runner == "auto" {
				configured = append(configured, fmt.Sprintf("%s: auto", dname))
				continue
			}
			if _, err := exec.LookPath(runner); err != nil {
				missing = append(missing, fmt.Sprintf("%s: %s not found", dname, runner))
			} else {
				configured = append(configured, fmt.Sprintf("%s: %s", dname, runner))
			}
		}
	}

	if len(missing) > 0 {
		return fmt.Sprintf("%d missing: %v", len(missing), missing), false
	}
	if len(configured) == 0 {
		return "none configured", true
	}
	return fmt.Sprintf("%v", configured), true
}

func checkTemplates(app *App) (string, bool) {
	if app.Cfg == nil {
		return "no config loaded", false
	}

	var errors []string
	for clientName, client := range app.Cfg.Clients {
		for domainName, domain := range client.Domains {
			// Try compiling all template strings
			for k, v := range domain.Env {
				if _, err := template.New("").Parse(v); err != nil {
					errors = append(errors, fmt.Sprintf("%s/%s env.%s: %v", clientName, domainName, k, err))
				}
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Sprintf("%d errors: %v", len(errors), errors), false
	}
	return "all templates compile", true
}

func checkProviders(app *App) (string, bool) {
	if app.Cfg == nil {
		return "no config loaded", false
	}

	var providers []string
	if app.Cfg.Secrets.Providers.Env != nil && app.Cfg.Secrets.Providers.Env.Enabled {
		providers = append(providers, "env")
	}
	if app.Cfg.Secrets.Providers.Command != nil && app.Cfg.Secrets.Providers.Command.Enabled {
		for name := range app.Cfg.Secrets.Providers.Command.Commands {
			providers = append(providers, "command:"+name)
		}
	}

	if len(providers) == 0 {
		return "no secret providers configured", true
	}
	return fmt.Sprintf("%d providers: %v", len(providers), providers), true
}

// renderEffectiveForDoctor computes and renders the effective config for a domain,
// so doctor checks can validate resolved paths rather than raw templates.
func renderEffectiveForDoctor(cfg *config.Root, clientName string, client *config.Client, domainName string) (*config.EffectiveConfig, error) {
	domainCfg, err := resolve.ResolveDomainExtends(client, domainName)
	if err != nil {
		return nil, err
	}
	eff := config.ComputeEffective(cfg.Defaults, client.Defaults, domainCfg)

	funcMap := config.DefaultFuncMap()
	funcMap["secret"] = func(provider, path string) (string, error) {
		return "<secret>", nil
	}

	ctx := config.NewTemplateContext(clientName, client.Tags, domainName, eff)
	if err := config.RenderEffective(eff, ctx, funcMap); err != nil {
		return nil, err
	}
	return eff, nil
}

func checkAuditFile(app *App) (string, bool) {
	if app.Cfg == nil {
		return "no config loaded", false
	}

	pol := app.Cfg.Defaults.Policy
	if pol == nil || pol.Audit == nil || !pol.Audit.Enabled {
		return "audit not enabled", true
	}

	if pol.Audit.Sink == "stdout" || pol.Audit.Sink == "stderr" {
		return fmt.Sprintf("sink: %s", pol.Audit.Sink), true
	}

	path := config.ExpandTilde(pol.Audit.File)
	if path == "" {
		path = config.ExpandTilde("~/.local/state/rctl/audit.log")
	}

	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); err != nil {
		// Try to create it
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Sprintf("cannot create audit dir %s: %v", dir, err), false
		}
	}

	// Try to open/create the file
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Sprintf("cannot write to %s: %v", path, err), false
	}
	f.Close()

	return fmt.Sprintf("writable: %s", path), true
}
