package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func versionCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print rctl version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(app.Out, "rctl %s\n", app.Version)
			return nil
		},
	}
}
