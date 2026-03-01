package cli

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

func clientsCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "clients",
		Short: "List configured clients",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.Cfg == nil || len(app.Cfg.Clients) == 0 {
				fmt.Fprintln(app.Out, "No clients configured.")
				return nil
			}

			names := make([]string, 0, len(app.Cfg.Clients))
			for name := range app.Cfg.Clients {
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
