package cli

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/lavr/rctl/internal/config"
	"github.com/lavr/rctl/internal/resolve"
	"github.com/spf13/cobra"
)

func showCmd(app *App) *cobra.Command {
	var jsonFlag bool

	cmd := &cobra.Command{
		Use:   "show <client> <domain>",
		Short: "Show effective configuration for a client+domain",
		Args:  cobra.ExactArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			_ = preloadConfig(app)
			switch len(args) {
			case 0:
				return completeClients(app, toComplete)
			case 1:
				return completeDomains(app, args[0], toComplete)
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			clientName, client, err := resolve.FindClientByNameOrAlias(app.Cfg, args[0])
			if err != nil {
				return err
			}

			domainName, _, err := resolve.FindDomainByNameOrAlias(client, args[1])
			if err != nil {
				return err
			}

			domainCfg, err := resolve.ResolveDomainExtends(client, domainName)
			if err != nil {
				return err
			}

			eff := config.ComputeEffective(app.Cfg.Defaults, client.Defaults, domainCfg)

			// Render templates (secret as no-op since show is read-only)
			funcMap := config.DefaultFuncMap()
			funcMap["secret"] = func(provider, path string) (string, error) { return "", nil }
			ctx := config.NewTemplateContext(clientName, client.Tags, domainName, eff)
			if err := config.RenderEffective(eff, ctx, funcMap); err != nil {
				return fmt.Errorf("template rendering: %w", err)
			}

			if jsonFlag {
				return printEffectiveJSON(app, clientName, domainName, eff)
			}
			printEffective(app, clientName, domainName, eff)
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonFlag, "json", false, "output as JSON")
	return cmd
}

type showOutput struct {
	Client       string                    `json:"client"`
	Domain       string                    `json:"domain"`
	Dir          string                    `json:"dir,omitempty"`
	Env          map[string]string         `json:"env,omitempty"`
	DefaultArgs  []string                  `json:"default_args,omitempty"`
	CommandsPath []string                  `json:"commands_path,omitempty"`
	PathAdd      []string                  `json:"path_add,omitempty"`
	Profiles     map[string]config.Profile `json:"profiles,omitempty"`
	Vars         map[string]string         `json:"vars,omitempty"`
	TaskRunner   string                    `json:"task_runner,omitempty"`
}

func printEffectiveJSON(app *App, clientName, domainName string, eff *config.EffectiveConfig) error {
	out := showOutput{
		Client:       clientName,
		Domain:       domainName,
		Dir:          eff.Dir,
		Env:          eff.Env,
		DefaultArgs:  eff.DefaultArgs,
		CommandsPath: eff.CommandsPath,
		PathAdd:      eff.PathAdd,
		Profiles:     eff.Profiles,
		Vars:         eff.Vars,
		TaskRunner:   eff.TaskRunner,
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	fmt.Fprintln(app.Out, string(data))
	return nil
}

func printEffective(app *App, clientName, domainName string, eff *config.EffectiveConfig) {
	fmt.Fprintf(app.Out, "Client: %s\n", clientName)
	fmt.Fprintf(app.Out, "Domain: %s\n", domainName)

	if eff.Dir != "" {
		fmt.Fprintf(app.Out, "Dir: %s\n", eff.Dir)
	}

	if len(eff.Env) > 0 {
		fmt.Fprintln(app.Out, "Env:")
		keys := sortedKeys(eff.Env)
		for _, k := range keys {
			fmt.Fprintf(app.Out, "  %s=%s\n", k, eff.Env[k])
		}
	}

	if len(eff.DefaultArgs) > 0 {
		fmt.Fprintf(app.Out, "Default args: %s\n", strings.Join(eff.DefaultArgs, " "))
	}

	if len(eff.CommandsPath) > 0 {
		fmt.Fprintln(app.Out, "Commands path:")
		for _, p := range eff.CommandsPath {
			fmt.Fprintf(app.Out, "  %s\n", p)
		}
	}

	if len(eff.PathAdd) > 0 {
		fmt.Fprintln(app.Out, "Path add:")
		for _, p := range eff.PathAdd {
			fmt.Fprintf(app.Out, "  %s\n", p)
		}
	}

	if eff.TaskRunner != "" {
		fmt.Fprintf(app.Out, "Task runner: %s\n", eff.TaskRunner)
	}

	if len(eff.Profiles) > 0 {
		fmt.Fprintln(app.Out, "Profiles:")
		names := make([]string, 0, len(eff.Profiles))
		for name := range eff.Profiles {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			p := eff.Profiles[name]
			if len(p.Args) > 0 {
				fmt.Fprintf(app.Out, "  %s -> %s %s\n", name, p.Cmd, strings.Join(p.Args, " "))
			} else {
				fmt.Fprintf(app.Out, "  %s -> %s\n", name, p.Cmd)
			}
		}
	}
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
