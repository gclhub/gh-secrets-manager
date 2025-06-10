package main

import (
	"testing"
)

func TestNewRootCmd_InvalidCommand(t *testing.T) {
	cmd := newRootCmd()

	// Test invalid command
	cmd.SetArgs([]string{"invalid-command"})
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid command but got none")
	}
	if err.Error() != `unknown command "invalid-command" for "secrets-manager"` {
		t.Errorf("Expected specific error message for invalid command, got: %v", err)
	}
}

func TestNewRootCmd_InvalidFlag(t *testing.T) {
	cmd := newRootCmd()

	// Test invalid flag
	cmd.SetArgs([]string{"--invalid-flag"})
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid flag but got none")
	}
	if err.Error() != "unknown flag: --invalid-flag" {
		t.Errorf("Expected specific error message for invalid flag, got: %v", err)
	}
}

// TODO: Add additional edge case and flag tests:
// - Test combination of valid and invalid flags
// - Test malformed flag syntax (e.g., --flag=)
// - Test long/short flag variants
// - Test flag value validation for different data types
// - Test subcommand-specific flag validation
// - Test help flag behavior
// - Test version flag behavior
// - Test verbose flag functionality