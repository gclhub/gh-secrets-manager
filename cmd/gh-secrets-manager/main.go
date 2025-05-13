package main

import (
	"fmt"
	"os"

	"gh-secrets-manager/pkg/api"
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
	cmd := &cobra.Command{
		Use:   "secrets-manager", // Removed gh prefix from here since it's in the template
		Short: "GitHub CLI extension for managing GitHub Actions secrets and variables",
		Long:  `A GitHub CLI extension for managing GitHub Actions secrets and variables, and Dependabot secrets at both organization and repository levels.`,
		Version: fmt.Sprintf("%s (%s)", version.Version, version.CommitHash),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

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

	// Initialize GitHub App client options
	opts := &api.ClientOptions{
		AuthMethod: api.AuthMethodPAT, // Default to PAT auth
	}

	// Add and initialize all commands
	cmd.AddCommand(newConfigCmd())
	addSecretCommands(cmd, opts)
	addVariableCommands(cmd, opts)
	addDependabotCommands(cmd, opts)

	return cmd
}
