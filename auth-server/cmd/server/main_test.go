package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gclhub/gh-secrets-manager/auth-server/pkg/auth"
)

func generateTestKey(t *testing.T) []byte {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	keyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	})

	return pemBytes
}

func generateTestHandler(t *testing.T) (*Handler, []byte) {
	t.Helper()
	privateKey := generateTestKey(t) // Use the same helper from auth_test.go
	
	return &Handler{
		privateKeyPEM: privateKey,
		organization:  "",
		team:          "",
		verbose:       false,
	}, privateKey
}

func TestHandleToken_OrganizationVerification(t *testing.T) {
	tests := []struct {
		name               string
		serverOrg          string
		queryParams        map[string]string
		mockGitHubResponses map[string]func(w http.ResponseWriter, r *http.Request)
		expectedStatus      int
		expectedError       string
	}{
		{
			name:      "No organization verification required",
			serverOrg: "",
			queryParams: map[string]string{
				"app-id":          "123456",
				"installation-id": "987654",
			},
			mockGitHubResponses: map[string]func(w http.ResponseWriter, r *http.Request){
				"/app/installations/987654/access_tokens": func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusCreated)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"token":      "ghs_test_token",
						"expires_at": time.Now().Add(time.Hour),
					})
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "Organization verification successful",
			serverOrg: "testorg",
			queryParams: map[string]string{
				"app-id":          "123456",
				"installation-id": "987654",
				"username":        "testuser",
			},
			mockGitHubResponses: map[string]func(w http.ResponseWriter, r *http.Request){
				"/app/installations/987654/access_tokens": func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusCreated)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"token":      "ghs_test_token",
						"expires_at": time.Now().Add(time.Hour),
					})
				},
				"/orgs/testorg/members/testuser": func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNoContent) // User is a member
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "Organization verification failed - not a member",
			serverOrg: "testorg",
			queryParams: map[string]string{
				"app-id":          "123456",
				"installation-id": "987654",
				"username":        "testuser",
			},
			mockGitHubResponses: map[string]func(w http.ResponseWriter, r *http.Request){
				"/app/installations/987654/access_tokens": func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusCreated)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"token":      "ghs_test_token",
						"expires_at": time.Now().Add(time.Hour),
					})
				},
				"/orgs/testorg/members/testuser": func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound) // User is not a member
				},
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "not a member of organization",
		},
		{
			name:      "Missing username when organization verification required",
			serverOrg: "testorg",
			queryParams: map[string]string{
				"app-id":          "123456",
				"installation-id": "987654",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "username query parameter is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, _ := generateTestHandler(t)
			handler.organization = tt.serverOrg

			// Create mock GitHub server
			var githubServer *httptest.Server
			if tt.mockGitHubResponses != nil {
				githubServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if mockFunc, exists := tt.mockGitHubResponses[r.URL.Path]; exists {
						mockFunc(w, r)
					} else {
						w.WriteHeader(http.StatusNotFound)
					}
				}))
				defer githubServer.Close()

				// Configure auth package to use mock server
				originalURL := auth.GetGitHubAPIBaseURL()
				auth.SetGitHubAPIBaseURL(githubServer.URL)
				defer auth.SetGitHubAPIBaseURL(originalURL)
			}

			// Create request
			reqURL := "/token"
			if len(tt.queryParams) > 0 {
				params := url.Values{}
				for k, v := range tt.queryParams {
					params.Add(k, v)
				}
				reqURL += "?" + params.Encode()
			}

			req := httptest.NewRequest(http.MethodPost, reqURL, nil)
			w := httptest.NewRecorder()

			// Call handler
			handler.handleToken(w, req)

			// Check response
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d but got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				body := w.Body.String()
				if !contains(body, tt.expectedError) {
					t.Errorf("Expected error containing %q but got %q", tt.expectedError, body)
				}
			}

			if tt.expectedStatus == http.StatusOK {
				// Verify we got a valid token response
				var tokenResp auth.TokenResponse
				if err := json.NewDecoder(w.Body).Decode(&tokenResp); err != nil {
					t.Errorf("Failed to decode token response: %v", err)
				}
				if tokenResp.Token == "" {
					t.Error("Expected non-empty token")
				}
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(substr) <= len(s) && s[len(s)-len(substr):] == substr) || 
		(len(substr) <= len(s) && s[:len(substr)] == substr) ||
		(len(substr) < len(s) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}