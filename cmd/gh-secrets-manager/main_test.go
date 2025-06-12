package main

import (
	"strings"
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

func TestNewRootCmd_FlagCombinations(t *testing.T) {
	cmd := newRootCmd()

	// Test combination of valid and invalid flags
	cmd.SetArgs([]string{"--verbose", "--invalid-flag"})
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid flag combination but got none")
	}
	if !strings.Contains(err.Error(), "unknown flag: --invalid-flag") {
		t.Errorf("Expected error message about invalid flag, got: %v", err)
	}
}

func TestNewRootCmd_MalformedFlagSyntax(t *testing.T) {
	cmd := newRootCmd()

	// Test malformed flag syntax (e.g., --flag=)
	cmd.SetArgs([]string{"--verbose="})
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for malformed flag syntax but got none")
	}
}

func TestNewRootCmd_LongShortFlagVariants(t *testing.T) {
	// Test short flag variant
	t.Run("Short verbose flag", func(t *testing.T) {
		cmd := newRootCmd()
		cmd.SetArgs([]string{"-v"})
		// Should not error, just show help
		err := cmd.Execute()
		if err != nil {
			t.Errorf("Unexpected error with short verbose flag: %v", err)
		}
	})

	// Test long flag variant
	t.Run("Long verbose flag", func(t *testing.T) {
		cmd := newRootCmd()
		cmd.SetArgs([]string{"--verbose"})
		// Should not error, just show help
		err := cmd.Execute()
		if err != nil {
			t.Errorf("Unexpected error with long verbose flag: %v", err)
		}
	})
}

func TestNewRootCmd_HelpFlag(t *testing.T) {
	cmd := newRootCmd()

	// Test help flag behavior
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Help flag should not return error, got: %v", err)
	}

	// Test short help flag
	cmd = newRootCmd()
	cmd.SetArgs([]string{"-h"})
	err = cmd.Execute()
	if err != nil {
		t.Errorf("Short help flag should not return error, got: %v", err)
	}
}

func TestNewRootCmd_VersionFlag(t *testing.T) {
	cmd := newRootCmd()

	// Test version flag behavior
	cmd.SetArgs([]string{"--version"})
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Version flag should not return error, got: %v", err)
	}
}

func TestNewRootCmd_SubcommandValidation(t *testing.T) {
	// Test that subcommands exist
	cmd := newRootCmd()
	
	// Debug: print all available commands
	t.Logf("Available commands:")
	for _, c := range cmd.Commands() {
		t.Logf("  - %s", c.Name())
	}
	
	// Test valid subcommands that we explicitly add
	validSubcommands := []string{"config", "secrets", "variables", "dependabot"}
	for _, subcmd := range validSubcommands {
		found := false
		for _, c := range cmd.Commands() {
			if c.Name() == subcmd {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand %q to exist", subcmd)
		}
	}

	// The completion and help commands are available in Cobra but may not show up in Commands()
	// Let's test that they're actually callable instead
	t.Run("completion command exists", func(t *testing.T) {
		cmd := newRootCmd()
		cmd.SetArgs([]string{"completion", "--help"})
		err := cmd.Execute()
		if err != nil {
			t.Errorf("Completion command should be available, got error: %v", err)
		}
	})

	t.Run("help command exists", func(t *testing.T) {
		cmd := newRootCmd()
		cmd.SetArgs([]string{"help"})
		err := cmd.Execute()
		if err != nil {
			t.Errorf("Help command should be available, got error: %v", err)
		}
	})
}

func TestNewRootCmd_VerboseFlag(t *testing.T) {
	cmd := newRootCmd()

	// Test that verbose flag is recognized and parsed
	cmd.SetArgs([]string{"--verbose"})
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Verbose flag should not cause error, got: %v", err)
	}

	// Test short verbose flag
	cmd = newRootCmd()
	cmd.SetArgs([]string{"-v"})
	err = cmd.Execute()
	if err != nil {
		t.Errorf("Short verbose flag should not cause error, got: %v", err)
	}
}