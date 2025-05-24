package main

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestNewSecretsCmd(t *testing.T) {
	cmd := newSecretsCmd()
	
	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}
	
	if cmd.Use != "secrets" {
		t.Errorf("Expected use to be 'secrets', got %q", cmd.Use)
	}
	
	if cmd.Short == "" {
		t.Error("Expected short description, got empty string")
	}
}

func TestAddSecretCommands(t *testing.T) {
	rootCmd := &cobra.Command{Use: "test"}
	
	// Pass nil for client options
	addSecretCommands(rootCmd, nil)
	
	// Find the secrets command
	var secretsCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "secrets" {
			secretsCmd = cmd
			break
		}
	}
	
	if secretsCmd == nil {
		t.Fatal("Expected to find 'secrets' command")
	}
	
	// Check for the secrets subcommands under the secrets command
	subcommandNames := []string{"list", "set", "delete"}
	for _, name := range subcommandNames {
		found := false
		for _, cmd := range secretsCmd.Commands() {
			if cmd.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find '%s' subcommand", name)
		}
	}
	
	// Find the set command and check its flags
	var setCmd *cobra.Command
	for _, cmd := range secretsCmd.Commands() {
		if cmd.Name() == "set" {
			setCmd = cmd
			break
		}
	}
	
	if setCmd != nil {
		requiredFlags := []string{"org", "repo", "name", "value", "file"}
		for _, flagName := range requiredFlags {
			if setCmd.Flags().Lookup(flagName) == nil {
				t.Errorf("Expected to find '%s' flag on 'set' command", flagName)
			}
		}
	}
}

func TestSplitRepo(t *testing.T) {
	testCases := []struct {
		name          string
		repo          string
		expectedOwner string
		expectedRepo  string
	}{
		{
			name:          "Valid repository",
			repo:          "owner/repo",
			expectedOwner: "owner",
			expectedRepo:  "repo",
		},
		{
			name:          "Missing slash",
			repo:          "ownerrepo",
			expectedOwner: "",
			expectedRepo:  "ownerrepo",
		},
		{
			name:          "Empty string",
			repo:          "",
			expectedOwner: "",
			expectedRepo:  "",
		},
		{
			name:          "Too many slashes",
			repo:          "owner/repo/extra",
			expectedOwner: "",
			expectedRepo:  "owner/repo/extra",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			owner, repo := splitRepo(tc.repo)

			if owner != tc.expectedOwner {
				t.Errorf("Expected owner %q, got %q", tc.expectedOwner, owner)
			}
			if repo != tc.expectedRepo {
				t.Errorf("Expected repo %q, got %q", tc.expectedRepo, repo)
			}
		})
	}
}

func TestAddCommonFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	
	addCommonFlags(cmd)
	
	requiredFlags := []string{"org", "repo", "property", "prop_value"}
	for _, flagName := range requiredFlags {
		if cmd.Flags().Lookup(flagName) == nil {
			t.Errorf("Expected to find '%s' flag", flagName)
		}
	}
}