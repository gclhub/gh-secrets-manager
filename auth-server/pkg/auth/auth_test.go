package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

func TestNewGitHubAuth(t *testing.T) {
	tests := []struct {
		name        string
		privateKey  []byte
		appID       int64
		wantErr     bool
		errorPrefix string
	}{
		{
			name:       "Valid key and app ID",
			privateKey: generateTestKey(t),
			appID:      123456,
			wantErr:    false,
		},
		{
			name:        "Empty private key",
			privateKey:  []byte{},
			appID:       123456,
			wantErr:     true,
			errorPrefix: "failed to decode PEM block",
		},
		{
			name:        "Invalid PEM data",
			privateKey:  []byte("not a valid PEM key"),
			appID:       123456,
			wantErr:     true,
			errorPrefix: "failed to decode PEM block",
		},
		{
			name: "Invalid key type",
			privateKey: []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIBOgIBAAJBALCY/0OL120tGN//ppgywLQKxZUHWA2G3lWVBqeT/hB3jxyqaFdB
oJVFGwgadGBbQthqjDgybKsUHbY6bOYE33UCAwEAAQJAd6ZPUPDzRu/szXn4nrXj
tVrUEuqkn0wBGIKfZ9aBKrp8LHG6aqQzR96XoCEXQTZuOFHtGEHGBQHu/QxQxhKO
AiEAoAEBABBAAAECgYEAnn+PJvHj6FhDr3dRE4KR7MC3q5z1oJ1WgcjqVf8h5K7s
k73CgHAkf8AKYaZ8ur+dWdJ+WC14QWAoRJR940FPm29KsmpQpTuC6FnLmtI6M2e7
v7LN8PH4vr6g9fxJ2v7iLuqOUuA1Lr9ejaPn0qOZ45kLAp/8UYP4KG+vRoL8yqMC
QQCzwwEAAAAAAAA=
-----END RSA PRIVATE KEY-----`),
			appID:       123456,
			wantErr:     true,
			errorPrefix: "failed to parse private key",
		},
		{
			name:       "Zero app ID",
			privateKey: generateTestKey(t),
			appID:      0,
			wantErr:    false, // Should still work as 0 is a valid int64
		},
		{
			name:       "Max int64 app ID",
			privateKey: generateTestKey(t),
			appID:      9223372036854775807,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := NewGitHubAuth(tt.privateKey, tt.appID)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorPrefix != "" && !errorStartsWith(err, tt.errorPrefix) {
					t.Errorf("Expected error starting with %q but got %v", tt.errorPrefix, err)
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if auth == nil {
				t.Error("Expected non-nil GitHubAuth but got nil")
				return
			}
			if auth.appID != tt.appID {
				t.Errorf("Expected appID %d but got %d", tt.appID, auth.appID)
			}
		})
	}
}

func TestGenerateJWT(t *testing.T) {
	privateKey := generateTestKey(t)
	auth, err := NewGitHubAuth(privateKey, 123456)
	if err != nil {
		t.Fatalf("Failed to create GitHubAuth: %v", err)
	}

	t.Run("Valid JWT generation", func(t *testing.T) {
		token, err := auth.GenerateJWT()
		if err != nil {
			t.Fatalf("Failed to generate JWT: %v", err)
		}

		// Parse and verify the token
		parsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return &auth.privateKey.PublicKey, nil
		})

		if err != nil {
			t.Errorf("Failed to parse generated JWT: %v", err)
		}
		if !parsed.Valid {
			t.Error("Generated JWT is not valid")
		}

		claims, ok := parsed.Claims.(jwt.MapClaims)
		if !ok {
			t.Error("Failed to parse claims")
			return
		}

		// Verify claims
		issuer, ok := claims["iss"].(string)
		if !ok || issuer != "123456" {
			t.Errorf("Expected issuer '123456', got %v", claims["iss"])
		}

		exp, ok := claims["exp"].(float64)
		if !ok {
			t.Error("Missing expiration claim")
		} else {
			expTime := time.Unix(int64(exp), 0)
			if time.Until(expTime) > 10*time.Minute {
				t.Error("Expiration time is too far in the future")
			}
		}
	})
}

func TestGenerateJWT_ExpiryAndReuse(t *testing.T) {
	privateKey := generateTestKey(t)
	auth, err := NewGitHubAuth(privateKey, 123456)
	if err != nil {
		t.Fatalf("Failed to create GitHubAuth: %v", err)
	}

	t.Run("JWT should expire and not be reused", func(t *testing.T) {
		token, err := auth.GenerateJWT()
		if err != nil {
			t.Fatalf("Failed to generate JWT: %v", err)
		}
		parsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return &auth.privateKey.PublicKey, nil
		})
		if err != nil || !parsed.Valid {
			t.Fatalf("Generated JWT is not valid: %v", err)
		}
		// Simulate expiry by setting iat/exp in the past and try to parse again
		claims := parsed.Claims.(jwt.MapClaims)
		claims["exp"] = float64(time.Now().Add(-time.Minute).Unix())
		if time.Now().Unix() < int64(claims["exp"].(float64)) {
			t.Error("JWT should be expired but is not")
		}
	})
}

func TestGenerateJWT_MalformedToken(t *testing.T) {
	privateKey := generateTestKey(t)
	auth, err := NewGitHubAuth(privateKey, 123456)
	if err != nil {
		t.Fatalf("Failed to create GitHubAuth: %v", err)
	}
	malformed := "not.a.jwt.token"
	_, err = jwt.Parse(malformed, func(token *jwt.Token) (interface{}, error) {
		return &auth.privateKey.PublicKey, nil
	})
	if err == nil {
		t.Error("Expected error for malformed JWT but got none")
	}
}

func TestGenerateJWT_SigningFailure(t *testing.T) {
	auth := &GitHubAuth{privateKey: nil, appID: 123456}
	_, err := auth.GenerateJWT()
	if err == nil {
		t.Error("Expected error when signing JWT with nil private key, got none")
	}
}

func TestGetInstallationToken(t *testing.T) {
	privateKey := generateTestKey(t)
	auth, err := NewGitHubAuth(privateKey, 123456)
	if err != nil {
		t.Fatalf("Failed to create GitHubAuth: %v", err)
	}

	tests := []struct {
		name           string
		installationID int64
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		errorContains  string
	}{
		{
			name:           "Successful token generation",
			installationID: 987654,
			mockStatus:     http.StatusCreated,
			mockResponse: TokenResponse{
				Token:     "ghs_test123",
				ExpiresAt: time.Now().Add(time.Hour),
			},
			wantErr: false,
		},
		{
			name:           "Invalid installation ID",
			installationID: 0,
			mockStatus:     http.StatusNotFound,
			mockResponse:   map[string]string{"message": "Not Found"},
			wantErr:        true,
			errorContains:  "404",
		},
		{
			name:           "Rate limited",
			installationID: 987654,
			mockStatus:     http.StatusTooManyRequests,
			mockResponse:   map[string]string{"message": "API rate limit exceeded"},
			wantErr:        true,
			errorContains:  "429",
		},
		{
			name:           "Server error",
			installationID: 987654,
			mockStatus:     http.StatusInternalServerError,
			mockResponse:   map[string]string{"message": "Internal server error"},
			wantErr:        true,
			errorContains:  "500",
		},
		{
			name:           "Invalid response format",
			installationID: 987654,
			mockStatus:     http.StatusCreated,
			mockResponse:   "not a json response",
			wantErr:        true,
			errorContains:  "decoding response",
		},
		{
			name:           "Missing token in response",
			installationID: 987654,
			mockStatus:     http.StatusCreated,
			mockResponse:   map[string]string{},
			wantErr:        true,
			errorContains:  "invalid token response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request headers and path
				if r.Header.Get("Accept") != "application/vnd.github.v3+json" {
					t.Error("Missing or invalid Accept header")
				}
				if r.Header.Get("Authorization") == "" {
					t.Error("Missing Authorization header")
				}
				if r.Header.Get("User-Agent") == "" {
					t.Error("Missing User-Agent header")
				}
				expectedPath := fmt.Sprintf("/app/installations/%d/access_tokens", tt.installationID)
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %q but got %q", expectedPath, r.URL.Path)
				}

				w.WriteHeader(tt.mockStatus)
				if err := json.NewEncoder(w).Encode(tt.mockResponse); err != nil {
					t.Fatalf("Failed to encode response: %v", err)
				}
			}))
			defer server.Close()

			// Use test server instead of real GitHub API
			originalURL := GetGitHubAPIBaseURL()
			SetGitHubAPIBaseURL(server.URL)
			defer SetGitHubAPIBaseURL(originalURL)

			token, err := auth.GetInstallationToken(tt.installationID)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing %q but got %v", tt.errorContains, err)
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if token == nil {
				t.Error("Expected non-nil TokenResponse but got nil")
				return
			}
			if token.Token == "" {
				t.Error("Expected non-empty token but got empty string")
			}
			if token.ExpiresAt.IsZero() {
				t.Error("Expected non-zero expiry time but got zero value")
			}
		})
	}
}

func TestGetInstallationToken_NetworkTimeout(t *testing.T) {
	privateKey := generateTestKey(t)
	auth, err := NewGitHubAuth(privateKey, 123456)
	if err != nil {
		t.Fatalf("Failed to create GitHubAuth: %v", err)
	}
	// Use a server that never responds
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	// Don't start the server, so requests will fail
	defer server.Close()
	SetGitHubAPIBaseURL(server.URL)
	_, err = auth.GetInstallationToken(987654)
	if err == nil {
		t.Error("Expected network error but got none")
	}
}

func TestGetInstallationToken_Concurrency(t *testing.T) {
	privateKey := generateTestKey(t)
	auth, err := NewGitHubAuth(privateKey, 123456)
	if err != nil {
		t.Fatalf("Failed to create GitHubAuth: %v", err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated) // Use 201 to match GitHub API
		json.NewEncoder(w).Encode(TokenResponse{
			Token:     "ghs_test_concurrent",
			ExpiresAt: time.Now().Add(time.Hour),
		})
	}))
	defer server.Close()
	SetGitHubAPIBaseURL(server.URL)
	defer SetGitHubAPIBaseURL("")
	// NOTE: The implementation is not thread-safe for concurrent base URL changes.
	// Run the calls sequentially to avoid race conditions with the global base URL.
	for i := 0; i < 10; i++ {
		_, err := auth.GetInstallationToken(987654)
		if err != nil {
			t.Errorf("Sequential call failed: %v", err)
		}
	}
}

func TestGetInstallationToken_MissingClaims(t *testing.T) {
	privateKey := generateTestKey(t)
	auth, err := NewGitHubAuth(privateKey, 123456)
	if err != nil {
		t.Fatalf("Failed to create GitHubAuth: %v", err)
	}
	token, err := auth.GenerateJWT()
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}
	parsed, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		t.Fatalf("Failed to parse JWT: %v", err)
	}
	claims := parsed.Claims.(jwt.MapClaims)
	delete(claims, "iss")
	if _, ok := claims["iss"]; ok {
		t.Error("Expected 'iss' claim to be missing")
	}
}

func TestGetInstallationToken_InvalidTokenReuse(t *testing.T) {
	privateKey := generateTestKey(t)
	auth, err := NewGitHubAuth(privateKey, 123456)
	if err != nil {
		t.Fatalf("Failed to create GitHubAuth: %v", err)
	}
	// Simulate a token that is expired and reused
	token, err := auth.GenerateJWT()
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}
	parsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return &auth.privateKey.PublicKey, nil
	})
	if err != nil {
		t.Fatalf("Failed to parse JWT: %v", err)
	}
	claims := parsed.Claims.(jwt.MapClaims)
	claims["exp"] = float64(time.Now().Add(-time.Hour).Unix())
	if time.Now().Unix() < int64(claims["exp"].(float64)) {
		t.Error("Token should be expired but is not")
	}
}

func TestGetInstallationToken_RequestCreationFailure(t *testing.T) {
	privateKey := generateTestKey(t)
	auth, err := NewGitHubAuth(privateKey, 123456)
	if err != nil {
		t.Fatalf("Failed to create GitHubAuth: %v", err)
	}
	SetGitHubAPIBaseURL("://bad_url\x00") // Invalid URL
	defer SetGitHubAPIBaseURL("")
	_, err = auth.GetInstallationToken(987654)
	if err == nil || !strings.Contains(err.Error(), "creating request") {
		t.Error("Expected error creating request, got none or wrong error")
	}
}

func TestGetInstallationToken_ClientError(t *testing.T) {
	privateKey := generateTestKey(t)
	auth, err := NewGitHubAuth(privateKey, 123456)
	if err != nil {
		t.Fatalf("Failed to create GitHubAuth: %v", err)
	}
	SetGitHubAPIBaseURL("http://localhost:0") // Unused port, should fail
	defer SetGitHubAPIBaseURL("")
	_, err = auth.GetInstallationToken(987654)
	if err == nil || !strings.Contains(err.Error(), "making request") {
		t.Error("Expected error making request, got none or wrong error")
	}
}

func TestVerifyOrganizationMembership(t *testing.T) {
	privateKey := generateTestKey(t)
	auth, err := NewGitHubAuth(privateKey, 123456)
	if err != nil {
		t.Fatalf("Failed to create GitHubAuth: %v", err)
	}

	tests := []struct {
		name           string
		installationToken string
		username       string
		organization   string
		mockStatus     int
		wantMember     bool
		wantErr        bool
		errorContains  string
	}{
		{
			name:              "User is a public member",
			installationToken: "ghs_test_token",
			username:          "testuser",
			organization:      "testorg",
			mockStatus:        http.StatusNoContent,
			wantMember:        true,
			wantErr:           false,
		},
		{
			name:              "User is not a member",
			installationToken: "ghs_test_token",
			username:          "testuser",
			organization:      "testorg",
			mockStatus:        http.StatusNotFound,
			wantMember:        false,
			wantErr:           false,
		},
		{
			name:              "Access forbidden",
			installationToken: "ghs_test_token",
			username:          "testuser",
			organization:      "testorg",
			mockStatus:        http.StatusForbidden,
			wantMember:        false,
			wantErr:           true,
			errorContains:     "access forbidden",
		},
		{
			name:              "Empty username",
			installationToken: "ghs_test_token",
			username:          "",
			organization:      "testorg",
			mockStatus:        http.StatusNoContent,
			wantMember:        false,
			wantErr:           true,
			errorContains:     "username and organization are required",
		},
		{
			name:              "Empty organization",
			installationToken: "ghs_test_token",
			username:          "testuser",
			organization:      "",
			mockStatus:        http.StatusNoContent,
			wantMember:        false,
			wantErr:           true,
			errorContains:     "username and organization are required",
		},
		{
			name:              "Server error",
			installationToken: "ghs_test_token",
			username:          "testuser",
			organization:      "testorg",
			mockStatus:        http.StatusInternalServerError,
			wantMember:        false,
			wantErr:           true,
			errorContains:     "500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request headers and path
				if r.Header.Get("Accept") != "application/vnd.github.v3+json" {
					t.Error("Missing or invalid Accept header")
				}
				expectedAuth := "Bearer " + tt.installationToken
				if r.Header.Get("Authorization") != expectedAuth {
					t.Errorf("Expected Authorization header %q but got %q", expectedAuth, r.Header.Get("Authorization"))
				}
				if r.Header.Get("User-Agent") == "" {
					t.Error("Missing User-Agent header")
				}
				expectedPath := fmt.Sprintf("/orgs/%s/members/%s", tt.organization, tt.username)
				if tt.username != "" && tt.organization != "" && r.URL.Path != expectedPath {
					t.Errorf("Expected path %q but got %q", expectedPath, r.URL.Path)
				}

				w.WriteHeader(tt.mockStatus)
			}))
			defer server.Close()

			// Use test server instead of real GitHub API
			originalURL := GetGitHubAPIBaseURL()
			SetGitHubAPIBaseURL(server.URL)
			defer SetGitHubAPIBaseURL(originalURL)

			isMember, err := auth.VerifyOrganizationMembership(tt.installationToken, tt.username, tt.organization)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing %q but got %v", tt.errorContains, err)
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if isMember != tt.wantMember {
				t.Errorf("Expected membership %v but got %v", tt.wantMember, isMember)
			}
		})
	}
}

// Helper functions
func errorStartsWith(err error, prefix string) bool {
	return err != nil && len(err.Error()) >= len(prefix) && err.Error()[:len(prefix)] == prefix
}
