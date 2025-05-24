package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Mock out the command execution
func mockExecuteCommand(root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err = root.Execute()
	return buf.String(), err
}

// Helper function to test command execution
func TestCommandExecution(t *testing.T) {
	cmd := newRootCmd()
	
	// Test help command
	output, err := mockExecuteCommand(cmd, "--help")
	if err != nil {
		t.Errorf("Error executing help command: %v", err)
	}
	if !strings.Contains(output, "Usage:") {
		t.Errorf("Expected usage information in help output, got: %s", output)
	}
}

func TestSecretCommandFlags(t *testing.T) {
	cmd := newRootCmd()
	secretsCmd := findSubCommand(cmd, "secrets")
	
	// Test secrets list command flags
	listCmd := findSubCommand(secretsCmd, "list")
	if listCmd == nil {
		t.Fatal("Expected secrets list command to exist")
	}
	
	// Check that required flags exist
	orgFlag := listCmd.Flags().Lookup("org")
	repoFlag := listCmd.Flags().Lookup("repo")
	
	if orgFlag == nil {
		t.Error("Expected --org flag on secrets list command")
	}
	
	if repoFlag == nil {
		t.Error("Expected --repo flag on secrets list command")
	}
}

func TestDependabotCommandFlags(t *testing.T) {
	cmd := newRootCmd()
	depCmd := findSubCommand(cmd, "dependabot")
	
	if depCmd == nil {
		t.Fatal("Expected dependabot command to exist")
	}
	
	// Check that subcommands exist
	listCmd := findSubCommand(depCmd, "list")
	setCmd := findSubCommand(depCmd, "set")
	deleteCmd := findSubCommand(depCmd, "delete")
	
	if listCmd == nil {
		t.Error("Expected dependabot list command to exist")
	}
	
	if setCmd == nil {
		t.Error("Expected dependabot set command to exist")
	}
	
	if deleteCmd == nil {
		t.Error("Expected dependabot delete command to exist")
	}
}

func TestVariablesCommandFlags(t *testing.T) {
	cmd := newRootCmd()
	varsCmd := findSubCommand(cmd, "variables")
	
	if varsCmd == nil {
		t.Fatal("Expected variables command to exist")
	}
	
	// Check that subcommands exist
	listCmd := findSubCommand(varsCmd, "list")
	setCmd := findSubCommand(varsCmd, "set")
	deleteCmd := findSubCommand(varsCmd, "delete")
	
	if listCmd == nil {
		t.Error("Expected variables list command to exist")
	}
	
	if setCmd == nil {
		t.Error("Expected variables set command to exist")
	}
	
	if deleteCmd == nil {
		t.Error("Expected variables delete command to exist")
	}
}

func TestConfigCommandFlags(t *testing.T) {
	cmd := newRootCmd()
	configCmd := findSubCommand(cmd, "config")
	
	if configCmd == nil {
		t.Fatal("Expected config command to exist")
	}
	
	// Check that subcommands exist
	getCmd := findSubCommand(configCmd, "get")
	setCmd := findSubCommand(configCmd, "set")
	deleteCmd := findSubCommand(configCmd, "delete")
	viewCmd := findSubCommand(configCmd, "view")
	
	if getCmd == nil {
		t.Error("Expected config get command to exist")
	}
	
	if setCmd == nil {
		t.Error("Expected config set command to exist")
	}
	
	if deleteCmd == nil {
		t.Error("Expected config delete command to exist")
	}
	
	if viewCmd == nil {
		t.Error("Expected config view command to exist")
	}
}