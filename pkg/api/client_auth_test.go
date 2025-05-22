package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestEnsureValidToken_NonGitHubAppAuth(t *testing.T) {
	client := &Client{
		opts: &ClientOptions{
			AuthMethod: AuthMethodPAT,
		},
	}
	
	// This should return nil for PAT auth
	err := client.ensureValidToken()
	if err != nil {
		t.Errorf("ensureValidToken with PAT auth returned error: %v", err)
	}
}

func TestEnsureValidToken_ValidToken(t *testing.T) {
	client := &Client{
		opts: &ClientOptions{
			AuthMethod: AuthMethodGitHubApp,
		},
		// Set token expiry to 10 minutes in the future
		expiresAt: time.Now().Add(10 * time.Minute),
	}
	
	// This should return nil for a non-expired token
	err := client.ensureValidToken()
	if err != nil {
		t.Errorf("ensureValidToken with valid token returned error: %v", err)
	}
}

func TestEnsureValidToken_ExpiredToken(t *testing.T) {
	// Setup test server to return a fake token
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"token": "test-refreshed-token", "expires_at": "2099-01-01T00:00:00Z"}`))
	}))
	defer server.Close()
	
	client := &Client{
		opts: &ClientOptions{
			AuthMethod: AuthMethodGitHubApp,
			AuthServer: server.URL,
			AppID:      12345,
			InstallationID: 67890,
		},
		// Set token expiry to the past
		expiresAt: time.Now().Add(-10 * time.Minute),
	}
	
	// This should trigger a token refresh
	err := client.ensureValidToken()
	if err != nil {
		t.Errorf("ensureValidToken with expired token returned error: %v", err)
	}
	
	// Check that the token was refreshed
	if client.authToken != "test-refreshed-token" {
		t.Errorf("Token was not refreshed, got: %q", client.authToken)
	}
	
	// Check that expiry time was updated
	if client.expiresAt.Before(time.Now()) {
		t.Errorf("Token expiry time was not updated correctly")
	}
}
