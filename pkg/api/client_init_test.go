package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClientWithOptions_GitHubAppAuth(t *testing.T) {
	// Setup test server to return a fake token
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"token": "test-token", "expires_at": "2099-01-01T00:00:00Z"}`))
	}))
	defer server.Close()
	
	// Test creating a client with GitHub App auth
	client, err := NewClientWithOptions(&ClientOptions{
		AuthMethod:     AuthMethodGitHubApp,
		AuthServer:     server.URL,
		AppID:          12345,
		InstallationID: 67890,
	})
	if err != nil {
		t.Fatalf("NewClientWithOptions with GitHub App auth returned error: %v", err)
	}
	
	if client == nil {
		t.Fatal("NewClientWithOptions with GitHub App auth returned nil client")
	}
	
	if client.opts.AuthMethod != AuthMethodGitHubApp {
		t.Errorf("Expected AuthMethod GitHubApp, got %v", client.opts.AuthMethod)
	}
	
	if client.authToken != "test-token" {
		t.Errorf("Expected auth token 'test-token', got %q", client.authToken)
	}
}

func TestNewClientWithOptions_UnsupportedAuth(t *testing.T) {
	// Test creating a client with unsupported auth method
	_, err := NewClientWithOptions(&ClientOptions{
		AuthMethod: AuthMethod(999), // Invalid auth method
	})
	
	if err == nil {
		t.Fatal("NewClientWithOptions with unsupported auth should return error")
	}
}

func TestRefreshToken_ErrorCases(t *testing.T) {
	// Test missing auth server URL
	client := &Client{
		opts: &ClientOptions{
			AuthMethod: AuthMethodGitHubApp,
			// No AuthServer URL
		},
	}
	
	err := client.refreshToken()
	if err == nil {
		t.Error("refreshToken with no auth server URL should return error")
	}
	
	// Test auth server returning non-200 status
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error"))
	}))
	defer errorServer.Close()
	
	client = &Client{
		opts: &ClientOptions{
			AuthMethod:     AuthMethodGitHubApp,
			AuthServer:     errorServer.URL,
			AppID:          12345,
			InstallationID: 67890,
		},
	}
	
	err = client.refreshToken()
	if err == nil {
		t.Error("refreshToken with server error should return error")
	}
	
	// Test auth server returning invalid JSON
	invalidJSONServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer invalidJSONServer.Close()
	
	client = &Client{
		opts: &ClientOptions{
			AuthMethod:     AuthMethodGitHubApp,
			AuthServer:     invalidJSONServer.URL,
			AppID:          12345,
			InstallationID: 67890,
		},
	}
	
	err = client.refreshToken()
	if err == nil {
		t.Error("refreshToken with invalid JSON response should return error")
	}
}
