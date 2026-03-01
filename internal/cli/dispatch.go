package cli

import (
	"fmt"
	"os"
	"os/user"
	"time"

	"github.com/lavr/rctl/internal/audit"
	"github.com/lavr/rctl/internal/config"
	"github.com/lavr/rctl/internal/policy"
	"github.com/lavr/rctl/internal/resolve"
	"github.com/lavr/rctl/internal/runner"
	"github.com/lavr/rctl/internal/secrets"
)

// runDispatch orchestrates the full pipeline:
// lookup -> extends -> effective -> secrets -> templates -> resolve -> policy -> run -> audit.
func runDispatch(app *App, clientName, domainName, cmdName string, userArgs []string) error {
	start := time.Now()

	// 1. Lookup client and domain (with alias support)
	resolvedClientName, client, err := resolve.FindClientByNameOrAlias(app.Cfg, clientName)
	if err != nil {
		return err
	}
	app.Verbose.Printf("client: %s", resolvedClientName)

	resolvedDomainName, _, err := resolve.FindDomainByNameOrAlias(client, domainName)
	if err != nil {
		return err
	}
	app.Verbose.Printf("domain: %s", resolvedDomainName)

	// 2. Resolve extends chain
	domainCfg, err := resolve.ResolveDomainExtends(client, resolvedDomainName)
	if err != nil {
		return err
	}

	// 3. Compute effective config
	eff := config.ComputeEffective(app.Cfg.Defaults, client.Defaults, domainCfg)

	// 4. Setup secrets registry and build FuncMap
	registry := setupSecrets(app.Cfg)
	funcMap := config.DefaultFuncMap()
	funcMap["secret"] = registry.SecretFunc()

	// 5. Render templates
	ctx := config.NewTemplateContext(resolvedClientName, client.Tags, resolvedDomainName, eff)
	if err := config.RenderEffective(eff, ctx, funcMap); err != nil {
		return fmt.Errorf("template rendering: %w", err)
	}

	app.Verbose.Printf("effective dir: %s", eff.Dir)
	app.Verbose.Printf("effective env keys: %v", mapKeys(eff.Env))

	// 6. Resolve command
	result, err := resolve.ResolveCommand(cmdName, userArgs, eff, app.Cfg.Builtins, app.Verbose)
	if err != nil {
		fmt.Fprintf(app.Err, "Error: %v\n", err)
		setExitCode(127)
		return nil
	}

	app.Verbose.Printf("resolved: %s (source: %s)", result.ExecPath, result.Source)

	// 7. Policy check
	effPolicy := computeEffectivePolicy(app.Cfg.Defaults.Policy, client.Defaults.Policy, domainCfg.Policy)
	if denial := policy.Check(effPolicy, cmdName, result.ExecPath); denial != nil {
		fmt.Fprintf(app.Err, "Policy denied: %s\n", denial.Reason)
		setExitCode(126)
		return nil
	}

	// 8. Setup audit logger
	auditLogger := setupAudit(effPolicy, app)
	defer auditLogger.Close()

	// 9. Run
	code := runner.Run(runner.Params{
		ExecPath:     result.ExecPath,
		Args:         result.Argv,
		Dir:          eff.Dir,
		Env:          eff.Env,
		PathAdd:      eff.PathAdd,
		DryRun:       app.DryRun,
		PrintCmdline: app.PrintCmdline,
		PrintEnv:     app.PrintEnv,
		Stdout:       app.Out,
		Stderr:       app.Err,
		Stdin:        os.Stdin,
	})

	// 10. Audit log
	logAuditEntry(auditLogger, effPolicy, resolvedClientName, resolvedDomainName,
		eff, result, code, time.Since(start))

	setExitCode(code)
	return nil
}

func setupSecrets(cfg *config.Root) *secrets.Registry {
	registry := secrets.NewRegistry()

	if cfg.Secrets.Providers.Env != nil && cfg.Secrets.Providers.Env.Enabled {
		registry.Register(&secrets.EnvProvider{})
	}

	if cfg.Secrets.Providers.Command != nil && cfg.Secrets.Providers.Command.Enabled {
		for _, p := range secrets.NewCommandProviders(cfg.Secrets.Providers.Command.Commands) {
			registry.Register(p)
		}
	}

	return registry
}

func computeEffectivePolicy(layers ...*config.PolicyConfig) *config.PolicyConfig {
	var result config.PolicyConfig
	for _, p := range layers {
		if p == nil {
			continue
		}
		result.DenyCommands = append(result.DenyCommands, p.DenyCommands...)
		if len(p.AllowCommands) > 0 {
			result.AllowCommands = p.AllowCommands // last non-empty wins
		}
		if len(p.AllowPathPrefixes) > 0 {
			result.AllowPathPrefixes = p.AllowPathPrefixes
		}
		if p.RequireTTY {
			result.RequireTTY = true
		}
		if p.Audit != nil {
			result.Audit = p.Audit
		}
	}
	return &result
}

func setupAudit(pol *config.PolicyConfig, app *App) audit.Logger {
	if pol == nil || pol.Audit == nil || !pol.Audit.Enabled {
		return &audit.NoopLogger{}
	}

	switch pol.Audit.Sink {
	case "stdout", "stderr":
		return audit.NewStderrLogger()
	case "file", "":
		path := config.ExpandTilde(pol.Audit.File)
		if path == "" {
			path = config.ExpandTilde("~/.local/state/rctl/audit.log")
		}
		logger, err := audit.NewFileLogger(path)
		if err != nil {
			// Warn but don't abort
			fmt.Fprintf(app.Err, "Warning: cannot open audit log %s: %v\n", path, err)
			return &audit.NoopLogger{}
		}
		return logger
	default:
		return &audit.NoopLogger{}
	}
}

func logAuditEntry(logger audit.Logger, pol *config.PolicyConfig, clientName, domainName string,
	eff *config.EffectiveConfig, result *resolve.Result, exitCode int, duration time.Duration) {

	entry := audit.NewEntry()
	entry.Client = clientName
	entry.Domain = domainName
	entry.Cwd = eff.Dir
	entry.ExecPath = result.ExecPath
	entry.ExitCode = exitCode
	entry.DurationMs = duration.Milliseconds()

	if u, err := user.Current(); err == nil {
		entry.User = u.Username
	}

	// Redact sensitive data
	var redactPatterns []string
	if pol != nil && pol.Audit != nil {
		redactPatterns = pol.Audit.RedactEnvKeys
	}

	entry.Argv = policy.RedactArgv(result.Argv)
	entry.EnvChanged = policy.RedactEnv(eff.Env, redactPatterns)

	_ = logger.Log(entry)
}

func mapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
