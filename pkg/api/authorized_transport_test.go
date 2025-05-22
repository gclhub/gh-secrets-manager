package api

import (
	"net/http"
	"testing"
)

// Mock transport that doesn't actually make HTTP calls
type noopTransport struct{}

func (t *noopTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Just return a minimal response without making a real HTTP call
	return &http.Response{StatusCode: 200}, nil
}

func TestAuthorizedTransport(t *testing.T) {
	// Create a request
	req, _ := http.NewRequest("GET", "https://api.github.com/test", nil)
	
	// Create the authorizedTransport with our noop transport as the underlying one
	origTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = origTransport }()
	http.DefaultTransport = &noopTransport{}
	
	transport := &authorizedTransport{token: "test-token"}
	
	// Call RoundTrip which will set the headers
	_, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
	
	// Check Authorization header
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		t.Errorf("Authorization header not set")
	}
	
	// Check Accept header
	acceptHeader := req.Header.Get("Accept")
	if acceptHeader != "application/vnd.github.v3+json" {
		t.Errorf("Accept header = %q, want %q", acceptHeader, "application/vnd.github.v3+json")
	}
	
	// Check User-Agent is set when not provided
	userAgent := req.Header.Get("User-Agent")
	if userAgent != "gh-secrets-manager" {
		t.Errorf("User-Agent header = %q, want %q", userAgent, "gh-secrets-manager")
	}
	
	// Test that existing User-Agent is preserved
	req2, _ := http.NewRequest("GET", "https://api.github.com/test", nil)
	req2.Header.Set("User-Agent", "existing-agent")
	
	_, err = transport.RoundTrip(req2)
	if err != nil {
		t.Fatalf("RoundTrip with custom User-Agent failed: %v", err)
	}
	
	// Check User-Agent is preserved
	userAgent = req2.Header.Get("User-Agent")
	if userAgent != "existing-agent" {
		t.Errorf("User-Agent header = %q, want %q", userAgent, "existing-agent")
	}
}
