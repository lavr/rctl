package cli

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lavr/rctl/internal/config"
	"github.com/lavr/rctl/internal/resolve"
	"github.com/spf13/cobra"
)

// completeClients returns client names, aliases, and subcommand names.
func completeClients(app *App, toComplete string) ([]string, cobra.ShellCompDirective) {
	if app.Cfg == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	seen := make(map[string]bool)
	var results []string

	for name, client := range app.Cfg.Clients {
		if !seen[name] {
			seen[name] = true
			results = append(results, name)
		}
		for _, alias := range client.Aliases {
			if !seen[alias] {
				seen[alias] = true
				results = append(results, alias)
			}
		}
	}

	sort.Strings(results)
	return results, cobra.ShellCompDirectiveNoFileComp
}

// completeDomains returns domain names and aliases for a given client.
func completeDomains(app *App, clientArg string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if app.Cfg == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	_, client, err := resolve.FindClientByNameOrAlias(app.Cfg, clientArg)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	seen := make(map[string]bool)
	var results []string

	for name, domain := range client.Domains {
		if !seen[name] {
			seen[name] = true
			results = append(results, name)
		}
		for _, alias := range domain.Aliases {
			if !seen[alias] {
				seen[alias] = true
				results = append(results, alias)
			}
		}
	}

	sort.Strings(results)
	return results, cobra.ShellCompDirectiveNoFileComp
}

// completeCommands returns profile names, builtin names, and executables from eff.Dir and commands_path.
// When toComplete contains "/" (e.g. "./contrib/"), switches to file listing from eff.Dir.
func completeCommands(app *App, clientArg, domainArg, toComplete string) ([]string, cobra.ShellCompDirective) {
	if app.Cfg == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Path-like input: switch to file completion from eff.Dir
	if strings.Contains(toComplete, "/") {
		return completeArgs(app, clientArg, domainArg, toComplete)
	}

	seen := make(map[string]bool)
	var results []string

	add := func(name string) {
		if !seen[name] {
			seen[name] = true
			results = append(results, name)
		}
	}

	eff := computeEffectiveForCompletion(app, clientArg, domainArg)

	// Profiles
	if eff != nil {
		for name := range eff.Profiles {
			add(name)
		}
	}

	// Builtins
	if app.Cfg.Builtins != nil {
		for name := range app.Cfg.Builtins {
			add(name)
		}
	}

	// Executables from eff.Dir
	if eff != nil && eff.Dir != "" {
		for _, name := range listExecutables(eff.Dir) {
			add(name)
		}
	}

	// Executables from commands_path
	if eff != nil {
		for _, dir := range eff.CommandsPath {
			for _, name := range listExecutables(dir) {
				add(name)
			}
		}
	}

	sort.Strings(results)
	return results, cobra.ShellCompDirectiveNoFileComp
}

// completeArgs returns files from eff.Dir matching the given prefix.
func completeArgs(app *App, clientArg, domainArg, toComplete string) ([]string, cobra.ShellCompDirective) {
	if app.Cfg == nil {
		return nil, cobra.ShellCompDirectiveDefault
	}

	eff := computeEffectiveForCompletion(app, clientArg, domainArg)
	if eff == nil || eff.Dir == "" {
		return nil, cobra.ShellCompDirectiveDefault
	}

	results := listFiles(eff.Dir, toComplete)
	if len(results) == 0 {
		return nil, cobra.ShellCompDirectiveDefault
	}

	// If single result is a directory, don't add space so user continues typing
	if len(results) == 1 && strings.HasSuffix(results[0], "/") {
		return results, cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
	}
	return results, cobra.ShellCompDirectiveNoFileComp
}

// computeEffectiveForCompletion computes the effective config for completion.
// Returns nil on any error (safe for completion context).
func computeEffectiveForCompletion(app *App, clientArg, domainArg string) *config.EffectiveConfig {
	if app.Cfg == nil {
		return nil
	}

	resolvedClientName, client, err := resolve.FindClientByNameOrAlias(app.Cfg, clientArg)
	if err != nil {
		return nil
	}

	resolvedDomainName, _, err := resolve.FindDomainByNameOrAlias(client, domainArg)
	if err != nil {
		return nil
	}

	domainCfg, err := resolve.ResolveDomainExtends(client, resolvedDomainName)
	if err != nil {
		return nil
	}

	eff := config.ComputeEffective(app.Cfg.Defaults, client.Defaults, domainCfg)

	funcMap := config.DefaultFuncMap()
	funcMap["secret"] = func(provider, path string) (string, error) {
		return "", nil
	}

	ctx := config.NewTemplateContext(resolvedClientName, client.Tags, resolvedDomainName, eff)
	if err := config.RenderEffective(eff, ctx, funcMap); err != nil {
		return nil
	}

	return eff
}

// listExecutables returns names of executable files in a directory.
func listExecutables(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.Mode()&0111 != 0 {
			names = append(names, e.Name())
		}
	}
	return names
}

// listFiles returns files and directories in dir matching the given prefix.
// Paths are relative to dir. Directories get a "/" suffix.
func listFiles(dir, prefix string) []string {
	// Split prefix into subdirectory and name prefix
	searchDir := dir
	namePrefix := prefix
	if i := strings.LastIndex(prefix, "/"); i >= 0 {
		searchDir = filepath.Join(dir, prefix[:i+1])
		namePrefix = prefix[i+1:]
	}

	entries, err := os.ReadDir(searchDir)
	if err != nil {
		return nil
	}

	var results []string
	for _, e := range entries {
		name := e.Name()
		// Skip hidden files unless prefix starts with .
		if strings.HasPrefix(name, ".") && !strings.HasPrefix(namePrefix, ".") {
			continue
		}
		if !strings.HasPrefix(name, namePrefix) {
			continue
		}

		rel := name
		if i := strings.LastIndex(prefix, "/"); i >= 0 {
			rel = prefix[:i+1] + name
		}

		if e.IsDir() {
			rel += "/"
		}
		results = append(results, rel)
	}
	return results
}
