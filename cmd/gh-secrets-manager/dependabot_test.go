package main

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestNewDependabotCmd(t *testing.T) {
	cmd := newDependabotCmd()
	
	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}
	
	if cmd.Use != "dependabot" {
		t.Errorf("Expected use to be 'dependabot', got %q", cmd.Use)
	}
	
	if cmd.Short == "" {
		t.Error("Expected short description, got empty string")
	}
}

func TestAddDependabotCommands(t *testing.T) {
	rootCmd := &cobra.Command{Use: "test"}
	
	addDependabotCommands(rootCmd, nil)
	
	// Find the dependabot command
	var dependabotCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "dependabot" {
			dependabotCmd = cmd
			break
		}
	}
	
	if dependabotCmd == nil {
		t.Fatal("Expected to find 'dependabot' command")
	}
	
	// Check for expected subcommands
	subcommandNames := []string{"list", "set", "delete"}
	for _, name := range subcommandNames {
		found := false
		for _, cmd := range dependabotCmd.Commands() {
			if cmd.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find '%s' subcommand", name)
		}
	}
}