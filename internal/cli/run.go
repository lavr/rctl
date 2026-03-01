package cli

import (
	"github.com/spf13/cobra"
)

func runCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:                "run <client> <domain> <cmd> [args...]",
		Short:              "Run a command in client+domain context (alias for positional dispatch)",
		Args:               cobra.MinimumNArgs(3),
		DisableFlagParsing: true,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			_ = preloadConfig(app)
			switch len(args) {
			case 0:
				return completeClients(app, toComplete)
			case 1:
				return completeDomains(app, args[0], toComplete)
			case 2:
				return completeCommands(app, args[0], args[1], toComplete)
			default:
				return completeArgs(app, args[0], args[1], toComplete)
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Re-parse our global flags from the args
			before, after := splitDashDash(args)
			if len(before) < 3 {
				return cmd.Help()
			}
			clientName := before[0]
			domainName := before[1]
			cmdName := before[2]
			userArgs := append(before[3:], after...)
			return runDispatch(app, clientName, domainName, cmdName, userArgs)
		},
	}
}
