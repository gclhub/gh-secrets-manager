package main

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestNewVariablesCmd(t *testing.T) {
	cmd := newVariablesCmd()
	
	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}
	
	if cmd.Use != "variables" {
		t.Errorf("Expected use to be 'variables', got %q", cmd.Use)
	}
	
	if cmd.Short == "" {
		t.Error("Expected short description, got empty string")
	}
}

func TestAddVariableCommands(t *testing.T) {
	rootCmd := &cobra.Command{Use: "test"}
	
	addVariableCommands(rootCmd, nil)
	
	// Find the variables command
	var variablesCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "variables" {
			variablesCmd = cmd
			break
		}
	}
	
	if variablesCmd == nil {
		t.Fatal("Expected to find 'variables' command")
	}
	
	// Check for expected subcommands
	subcommandNames := []string{"list", "set", "delete"}
	for _, name := range subcommandNames {
		found := false
		for _, cmd := range variablesCmd.Commands() {
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