package cli

import (
	"encoding/json"
	"fmt"

	"github.com/lavr/rctl/internal/config"
	"github.com/lavr/rctl/internal/resolve"
	"github.com/spf13/cobra"
)

func whichCmd(app *App) *cobra.Command {
	var jsonFlag bool

	cmd := &cobra.Command{
		Use:   "which <client> <domain> <cmd>",
		Short: "Show where a command resolves to",
		Args:  cobra.ExactArgs(3),
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

			result, err := resolve.ResolveCommand(args[2], nil, eff, app.Cfg.Builtins, app.Verbose)
			if err != nil {
				if jsonFlag {
					out := whichOutput{
						Command: args[2],
						Client:  clientName,
						Domain:  domainName,
						Found:   false,
					}
					data, _ := json.MarshalIndent(out, "", "  ")
					fmt.Fprintln(app.Out, string(data))
					return nil
				}
				fmt.Fprintf(app.Out, "%s: not found\n", args[2])
				fmt.Fprintf(app.Out, "Client: %s, Domain: %s\n", clientName, domainName)
				if eff.Dir != "" {
					fmt.Fprintf(app.Out, "Searched: %s (cwd)\n", eff.Dir)
				}
				for _, p := range eff.CommandsPath {
					fmt.Fprintf(app.Out, "Searched: %s (commands_path)\n", p)
				}
				fmt.Fprintln(app.Out, "Searched: PATH")
				return nil
			}

			if jsonFlag {
				out := whichOutput{
					Command:  args[2],
					Client:   clientName,
					Domain:   domainName,
					Found:    true,
					ExecPath: result.ExecPath,
					Source:   result.Source,
				}
				data, _ := json.MarshalIndent(out, "", "  ")
				fmt.Fprintln(app.Out, string(data))
				return nil
			}

			fmt.Fprintf(app.Out, "%s\n", result.ExecPath)
			fmt.Fprintf(app.Out, "Source: %s\n", result.Source)
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonFlag, "json", false, "output as JSON")
	return cmd
}

type whichOutput struct {
	Command  string `json:"command"`
	Client   string `json:"client"`
	Domain   string `json:"domain"`
	Found    bool   `json:"found"`
	ExecPath string `json:"exec_path,omitempty"`
	Source   string `json:"source,omitempty"`
}
