package cli

import (
	"fmt"
	"sort"

	"github.com/lavr/rctl/internal/resolve"
	"github.com/spf13/cobra"
)

func domainsCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "domains <client>",
		Short: "List domains for a client",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			_ = preloadConfig(app)
			if len(args) == 0 {
				return completeClients(app, toComplete)
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			_, client, err := resolve.FindClientByNameOrAlias(app.Cfg, args[0])
			if err != nil {
				return err
			}

			if len(client.Domains) == 0 {
				fmt.Fprintf(app.Out, "No domains configured for %s.\n", args[0])
				return nil
			}

			names := make([]string, 0, len(client.Domains))
			for name := range client.Domains {
				names = append(names, name)
			}
			sort.Strings(names)

			for _, name := range names {
				fmt.Fprintln(app.Out, name)
			}
			return nil
		},
	}
}
