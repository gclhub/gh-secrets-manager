package api

import (
	"os"
	"testing"
)

func TestNewClient_Minimal(t *testing.T) {
	// Save current environment and restore after test
	origToken := os.Getenv("GITHUB_TOKEN")
	defer os.Setenv("GITHUB_TOKEN", origToken)

	// Set a test token for this test
	os.Setenv("GITHUB_TOKEN", "test-token")

	client, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if client == nil {
		t.Fatal("Expected non-nil client")
	}
	if client.ctx == nil {
		t.Error("Expected non-nil context")
	}
	if client.github == nil {
		t.Error("Expected non-nil github client")
	}
	if client.opts == nil {
		t.Error("Expected non-nil options")
	}
	if client.opts.AuthMethod != AuthMethodPAT {
		t.Errorf("Expected auth method %v, got %v", AuthMethodPAT, client.opts.AuthMethod)
	}
}