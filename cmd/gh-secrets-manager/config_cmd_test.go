package main

import (
	"testing"
)

func TestNewConfigCmd(t *testing.T) {
	cmd := newConfigCmd()
	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}

	// Check basic properties
	if cmd.Use != "config" {
		t.Errorf("Expected Use to be 'config', got '%s'", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("Expected non-empty Short description")
	}

	// Check subcommands
	if len(cmd.Commands()) != 4 {
		t.Errorf("Expected 4 subcommands, got %d", len(cmd.Commands()))
	}

	// Check each subcommand exists
	subCmds := map[string]bool{
		"view":   false,
		"get":    false,
		"set":    false,
		"delete": false,
	}

	for _, sub := range cmd.Commands() {
		subCmds[sub.Name()] = true
	}

	for name, found := range subCmds {
		if !found {
			t.Errorf("Expected '%s' subcommand, not found", name)
		}
	}
}

func TestNewConfigViewCmd(t *testing.T) {
	cmd := newConfigViewCmd()
	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}

	// Check basic properties
	if cmd.Use != "view" {
		t.Errorf("Expected Use to be 'view', got '%s'", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("Expected non-empty Short description")
	}
	if cmd.RunE == nil {
		t.Error("Expected RunE function to be set")
	}
}

func TestNewConfigGetCmd(t *testing.T) {
	cmd := newConfigGetCmd()
	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}

	// Check basic properties
	if cmd.Use != "get <key>" {
		t.Errorf("Expected Use to be 'get <key>', got '%s'", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("Expected non-empty Short description")
	}
	if cmd.RunE == nil {
		t.Error("Expected RunE function to be set")
	}
}

func TestNewConfigSetCmd(t *testing.T) {
	cmd := newConfigSetCmd()
	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}

	// Check basic properties
	if cmd.Use != "set <key> <value>" {
		t.Errorf("Expected Use to be 'set <key> <value>', got '%s'", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("Expected non-empty Short description")
	}
	if cmd.RunE == nil {
		t.Error("Expected RunE function to be set")
	}
}

func TestNewConfigDeleteCmd(t *testing.T) {
	cmd := newConfigDeleteCmd()
	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}

	// Check basic properties
	if cmd.Use != "delete <key>" {
		t.Errorf("Expected Use to be 'delete <key>', got '%s'", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("Expected non-empty Short description")
	}
	if cmd.RunE == nil {
		t.Error("Expected RunE function to be set")
	}
}

func TestValidateURL(t *testing.T) {
	validURLs := []string{
		"https://example.com",
		"https://api.github.com",
		"http://localhost:8080",
		"http://127.0.0.1:3000",
	}

	invalidURLs := []string{
		"http://example.com",  // HTTP for non-localhost
		"ftp://example.com",   // Wrong scheme
		"https://",           // Incomplete
		"example.com",        // Missing scheme
	}

	for _, url := range validURLs {
		if err := validateURLValue(url); err != nil {
			t.Errorf("Expected valid URL for %q, got error: %v", url, err)
		}
	}

	for _, url := range invalidURLs {
		if err := validateURLValue(url); err == nil {
			t.Errorf("Expected error for invalid URL %q, got nil", url)
		}
	}
}