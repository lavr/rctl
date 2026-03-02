package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Execute creates the root command and runs it. Returns exit code.
func Execute(version string) int {
	app := NewApp(version)

	root := &cobra.Command{
		Use:           "rctl",
		Short:         "Run commands in client+domain context",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			app.SetupVerbose()
			// Skip config loading for commands that don't need it
			if cmd.Name() == "version" || cmd.Name() == "completion" {
				return nil
			}
			return app.LoadConfig()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			// Positional dispatch: rctl <client> <domain> <cmd> [args...]
			return dispatch(app, args)
		},
	}

	// Accept unknown flags for passthrough to child commands
	root.FParseErrWhitelist.UnknownFlags = true

	// Global flags
	root.PersistentFlags().StringVar(&app.ConfigPath, "config", "", "path to config file")
	root.PersistentFlags().BoolVar(&app.VerboseFlag, "verbose", false, "enable verbose logging to stderr")
	root.PersistentFlags().BoolVar(&app.DryRun, "dry-run", false, "print what would be executed without running")
	root.PersistentFlags().BoolVar(&app.PrintCmdline, "print-cmdline", false, "print the command line")
	root.PersistentFlags().BoolVar(&app.PrintEnv, "print-env", false, "print effective environment overrides")

	// Subcommands
	root.AddCommand(
		versionCmd(app),
		clientsCmd(app),
		domainsCmd(app),
		showCmd(app),
		whichCmd(app),
		policyCmd(app),
		doctorCmd(app),
		runCmd(app),
		completionCmd(root),
	)

	// Cobra doesn't natively support positional dispatch alongside subcommands.
	// We handle args that don't match a subcommand via Args/RunE.
	root.Args = cobra.ArbitraryArgs
	root.TraverseChildren = true

	// Dynamic shell completion for positional dispatch
	root.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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
	}

	if err := root.Execute(); err != nil {
		fmt.Fprintln(app.Err, "Error:", err)
		return 1
	}
	return exitCode
}

// exitCode is set by dispatch when it runs a child process.
var exitCode int

// setExitCode sets the exit code for the process.
func setExitCode(code int) {
	exitCode = code
}

// dispatch handles positional argument dispatch: rctl <client> <domain> <cmd> [args...]
func dispatch(app *App, args []string) error {
	// Need at least client, domain, cmd
	if len(args) < 3 {
		return fmt.Errorf("usage: rctl <client> <domain> <cmd> [args...]\n\nRun 'rctl --help' for more information")
	}

	// Handle -- separator: find and split
	var rawArgs []string
	for i, a := range args {
		if a == "--" {
			rawArgs = args[i+1:]
			args = args[:i]
			break
		}
	}

	if len(args) < 3 {
		return fmt.Errorf("usage: rctl <client> <domain> <cmd> [args...]")
	}

	clientName := args[0]
	domainName := args[1]
	cmdName := args[2]
	userArgs := append(args[3:], rawArgs...)

	return runDispatch(app, clientName, domainName, cmdName, userArgs)
}

// splitDashDash splits args at -- into before and after.
func splitDashDash(args []string) (before, after []string) {
	for i, a := range args {
		if a == "--" {
			return args[:i], args[i+1:]
		}
	}
	return args, nil
}

// preloadConfig ensures config is loaded (for commands that bypass PersistentPreRunE).
func preloadConfig(app *App) error {
	if app.Cfg != nil {
		return nil
	}
	return app.LoadConfig()
}

func init() {
	// Reset exit code at start
	exitCode = 0
}
