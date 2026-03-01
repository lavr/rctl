package cli

import (
	"fmt"

	"github.com/lavr/rctl/internal/config"
	"github.com/lavr/rctl/internal/policy"
	"github.com/lavr/rctl/internal/resolve"
	"github.com/spf13/cobra"
)

func policyCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "policy <client> <domain> <cmd>",
		Short: "Check if a command is allowed by policy",
		Args:  cobra.MinimumNArgs(3),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			_ = preloadConfig(app)
			switch len(args) {
			case 0:
				return completeClients(app, toComplete)
			case 1:
				return completeDomains(app, args[0], toComplete)
			case 2:
				return completeCommands(app, args[0], args[1], toComplete)
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

			cmdName := args[2]
			result, err := resolve.ResolveCommand(cmdName, nil, eff, app.Cfg.Builtins, app.Verbose)
			if err != nil {
				fmt.Fprintf(app.Out, "Command %q: not found (would be exit 127)\n", cmdName)
				return nil
			}

			effPolicy := computeEffectivePolicy(app.Cfg.Defaults.Policy, client.Defaults.Policy, domainCfg.Policy)
			denial := policy.Check(effPolicy, cmdName, result.ExecPath)

			fmt.Fprintf(app.Out, "Client: %s\n", clientName)
			fmt.Fprintf(app.Out, "Domain: %s\n", domainName)
			fmt.Fprintf(app.Out, "Command: %s\n", cmdName)
			fmt.Fprintf(app.Out, "Executable: %s\n", result.ExecPath)

			if denial != nil {
				fmt.Fprintf(app.Out, "Result: DENIED\n")
				fmt.Fprintf(app.Out, "Reason: %s\n", denial.Reason)
			} else {
				fmt.Fprintf(app.Out, "Result: ALLOWED\n")
			}

			// Show policy details
			if effPolicy != nil {
				if len(effPolicy.AllowCommands) > 0 {
					fmt.Fprintf(app.Out, "Allow commands: %v\n", effPolicy.AllowCommands)
				}
				if len(effPolicy.DenyCommands) > 0 {
					fmt.Fprintf(app.Out, "Deny commands: %v\n", effPolicy.DenyCommands)
				}
				if len(effPolicy.AllowPathPrefixes) > 0 {
					fmt.Fprintf(app.Out, "Allow path prefixes: %v\n", effPolicy.AllowPathPrefixes)
				}
				if effPolicy.RequireTTY {
					fmt.Fprintln(app.Out, "Require TTY: true")
				}
			}

			return nil
		},
	}
}
