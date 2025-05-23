package main

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestVerboseFlag(t *testing.T) {
	// Create a new root command
	cmd := newRootCmd()

	// Check if the verbose flag is set to false by default
	verboseFlag := cmd.PersistentFlags().Lookup("verbose")
	if verboseFlag == nil {
		t.Fatal("Expected verbose flag to exist")
	}

	// The flag should be a boolean flag
	if verboseFlag.Value.Type() != "bool" {
		t.Errorf("Expected verbose flag to be a boolean, got %s", verboseFlag.Value.Type())
	}

	// The default value should be false
	defaultValue := verboseFlag.Value.String()
	if defaultValue != "false" {
		t.Errorf("Expected verbose flag default value to be 'false', got '%s'", defaultValue)
	}
}

func TestAllCommandsHaveRun(t *testing.T) {
	// Create a new root command
	cmd := newRootCmd()
	
	// Function to recursively check all commands have a Run or RunE function
	var checkCommand func(cmd *cobra.Command, path string)
	checkCommand = func(cmd *cobra.Command, path string) {
		// Skip commands with template markup
		if cmd.Use == "config get <key>" || cmd.Use == "config set <key> <value>" || cmd.Use == "config delete <key>" {
			return
		}

		// Check that leaf commands (commands without subcommands) have a Run or RunE function
		if len(cmd.Commands()) == 0 && !cmd.HasSubCommands() && !cmd.IsAdditionalHelpTopicCommand() {
			if cmd.Run == nil && cmd.RunE == nil {
				t.Errorf("Command '%s' has no Run or RunE function", path)
			}
		}
		
		// Recursively check subcommands
		for _, subCmd := range cmd.Commands() {
			subPath := path + " " + subCmd.Name()
			checkCommand(subCmd, subPath)
		}
	}
	
	// Start the recursive check from the root command
	checkCommand(cmd, cmd.Name())
}

// TestCommandHelp verifies that all commands have a short and long help text
func TestCommandHelp(t *testing.T) {
	// Create a new root command
	cmd := newRootCmd()
	
	// Function to recursively check all commands have help text
	var checkCommand func(cmd *cobra.Command, path string)
	checkCommand = func(cmd *cobra.Command, path string) {
		// Check that all commands have a short help text
		if cmd.Short == "" {
			t.Errorf("Command '%s' has no short help text", path)
		}
		
		// Check that all commands have a long help text
		if cmd.Long == "" {
			t.Errorf("Command '%s' has no long help text", path)
		}
		
		// Recursively check subcommands
		for _, subCmd := range cmd.Commands() {
			subPath := path + " " + subCmd.Name()
			checkCommand(subCmd, subPath)
		}
	}
	
	// Start the recursive check from the root command
	checkCommand(cmd, cmd.Name())
}

func TestFindSubCommand(t *testing.T) {
	// Create a test command with subcommands
	rootCmd := &cobra.Command{Use: "root"}
	subCmd1 := &cobra.Command{Use: "sub1"}
	subCmd2 := &cobra.Command{Use: "sub2"}
	rootCmd.AddCommand(subCmd1, subCmd2)
	
	// Test finding an existing subcommand
	result := findSubCommand(rootCmd, "sub1")
	if result != subCmd1 {
		t.Errorf("Expected to find subcommand 'sub1', got %v", result)
	}
	
	// Test finding another existing subcommand
	result = findSubCommand(rootCmd, "sub2")
	if result != subCmd2 {
		t.Errorf("Expected to find subcommand 'sub2', got %v", result)
	}
	
	// Test finding a non-existent subcommand
	result = findSubCommand(rootCmd, "nonexistent")
	if result != nil {
		t.Errorf("Expected nil for non-existent subcommand, got %v", result)
	}
}