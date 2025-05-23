package main

import (
	"testing"
	
	"github.com/spf13/cobra"
)

func TestNewRootCmd(t *testing.T) {
	// Test that the root command is created successfully
	cmd := newRootCmd()
	if cmd == nil {
		t.Fatal("Expected non-nil command, got nil")
	}

	// Check command name and structure
	if cmd.Use != "secrets-manager" {
		t.Errorf("Expected command name to be 'secrets-manager', got '%s'", cmd.Use)
	}

	// Check that the command has subcommands
	if len(cmd.Commands()) < 3 { // should have at least config, secrets, and variables commands
		t.Errorf("Expected at least 3 subcommands, got %d", len(cmd.Commands()))
	}

	// Verify config command exists
	configCmd := findSubCommand(cmd, "config")
	if configCmd == nil {
		t.Error("Expected 'config' subcommand, not found")
	}

	// Verify secrets command exists
	secretsCmd := findSubCommand(cmd, "secrets")
	if secretsCmd == nil {
		t.Error("Expected 'secrets' subcommand, not found")
	}

	// Verify variables command exists
	variablesCmd := findSubCommand(cmd, "variables")
	if variablesCmd == nil {
		t.Error("Expected 'variables' subcommand, not found")
	}

	// Verify dependabot command exists
	dependabotCmd := findSubCommand(cmd, "dependabot")
	if dependabotCmd == nil {
		t.Error("Expected 'dependabot' subcommand, not found")
	}

	// Verify verbose flag exists
	verboseFlag := cmd.PersistentFlags().Lookup("verbose")
	if verboseFlag == nil {
		t.Error("Expected 'verbose' flag, not found")
	}
}

// Helper function to find a subcommand by name
func findSubCommand(cmd *cobra.Command, name string) *cobra.Command {
	for _, subCmd := range cmd.Commands() {
		if subCmd.Name() == name {
			return subCmd
		}
	}
	return nil
}