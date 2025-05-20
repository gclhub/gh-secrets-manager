package main

import (
	"fmt"
	"os"

	"gh-secrets-manager/pkg/api"
	"gh-secrets-manager/pkg/config"
	"gh-secrets-manager/pkg/version"

	"github.com/spf13/cobra"
)

func main() {
	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:     "secrets-manager",
		Short:   "GitHub CLI extension for managing GitHub Actions secrets and variables",
		Long:    `A GitHub CLI extension for managing GitHub Actions secrets and variables, and Dependabot secrets at both organization and repository levels.`,
		Version: fmt.Sprintf("%s (%s)", version.Version, version.CommitHash),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Set the API verbosity based on the flag
			api.Verbose = verbose
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// Add verbose flag to all commands
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")

	// Set custom usage template to ensure gh prefix appears everywhere
	cmd.SetUsageTemplate(`Usage:
  gh {{.UseLine}}{{if .HasAvailableSubCommands}}
  gh {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "gh {{.CommandPath}} [command] --help" for more information about a command.{{end}}
`)

	// Initialize client options - try GitHub App first, fall back to PAT
	cfg, err := config.Load()
	var opts *api.ClientOptions
	if err != nil || !cfg.IsGitHubAppConfigured() {
		opts = &api.ClientOptions{AuthMethod: api.AuthMethodPAT}
	} else {
		opts = &api.ClientOptions{
			AuthMethod:     api.AuthMethodGitHubApp,
			AppID:          cfg.AppID,
			InstallationID: cfg.InstallationID,
			AuthServer:     cfg.AuthServer,
		}
	}

	// Add and initialize all commands
	cmd.AddCommand(newConfigCmd())
	addSecretCommands(cmd, opts)
	addVariableCommands(cmd, opts)
	addDependabotCommands(cmd, opts)

	return cmd
}
