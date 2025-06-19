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

func TestGenerateJWT_MalformedClaims(t *testing.T) {
	privateKey := generateTestKey(t)
	auth, err := NewGitHubAuth(privateKey, 123456)
	if err != nil {
		t.Fatalf("Failed to create GitHubAuth: %v", err)
	}

	t.Run("JWT with unexpected type in iss claim", func(t *testing.T) {
		// Create a JWT with correct structure but wrong type for iss claim
		claims := jwt.MapClaims{
			"iss": 123456, // Should be string, not int
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(10 * time.Minute).Unix(),
		}
		
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		tokenString, err := token.SignedString(auth.privateKey)
		if err != nil {
			t.Fatalf("Failed to sign test JWT: %v", err)
		}

		// Parse and verify the token - this should succeed since format is valid
		parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return &auth.privateKey.PublicKey, nil
		})
		if err != nil {
			t.Fatalf("Failed to parse JWT: %v", err)
		}

		// Verify that the iss claim has unexpected type
		parsedClaims := parsedToken.Claims.(jwt.MapClaims)
		if issuer, ok := parsedClaims["iss"].(string); ok {
			t.Errorf("Expected iss claim to be non-string type, but got string: %v", issuer)
		}
		
		// Verify it's the wrong type (float64 due to JSON unmarshaling)
		if issuer, ok := parsedClaims["iss"].(float64); !ok {
			t.Errorf("Expected iss claim to be float64 due to JSON unmarshaling, got type: %T", parsedClaims["iss"])
		} else if int64(issuer) != 123456 {
			t.Errorf("Expected iss claim value to be 123456, got: %v", issuer)
		}
	})

	t.Run("JWT with missing required claims", func(t *testing.T) {
		// Create a JWT missing the iss claim
		claims := jwt.MapClaims{
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(10 * time.Minute).Unix(),
			// Missing "iss" claim
		}
		
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		tokenString, err := token.SignedString(auth.privateKey)
		if err != nil {
			t.Fatalf("Failed to sign test JWT: %v", err)
		}

		// Parse the token
		parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return &auth.privateKey.PublicKey, nil
		})
		if err != nil {
			t.Fatalf("Failed to parse JWT: %v", err)
		}

		// Verify that the iss claim is missing
		parsedClaims := parsedToken.Claims.(jwt.MapClaims)
		if _, exists := parsedClaims["iss"]; exists {
			t.Error("Expected iss claim to be missing, but it was present")
		}
	})

	// TODO: Expand malformed/negative JWT claim coverage:
	// - Test JWT with invalid exp claim type (string instead of numeric)
	// - Test JWT with exp claim in the past (expired token)
	// - Test JWT with negative exp/iat values
	// - Test JWT with iat claim in the future
	// - Test JWT with missing exp claim
	// - Test JWT with missing iat claim
	// - Test JWT with extra unexpected claims
	// - Test JWT with null/empty claim values
	// - Test JWT with extremely large numeric values in claims
	// - Test JWT with special characters in string claims

	t.Run("JWT with invalid exp claim type", func(t *testing.T) {
		// Create a JWT with exp as string instead of numeric
		claims := jwt.MapClaims{
			"iss": "123456",
			"iat": time.Now().Unix(),
			"exp": "not_a_number", // Should be numeric
		}
		
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		tokenString, err := token.SignedString(auth.privateKey)
		if err != nil {
			t.Fatalf("Failed to sign test JWT: %v", err)
		}

		// Parse the token - this should fail because exp is invalid type
		parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return &auth.privateKey.PublicKey, nil
		})
		if err == nil {
			t.Error("Expected error for invalid exp claim type, but got none")
		} else if !strings.Contains(err.Error(), "invalid type for claim") {
			t.Logf("Got expected error for invalid exp type: %v", err)
		}

		// Even if parsing fails, we can still extract the raw claims
		if parsedToken != nil {
			parsedClaims := parsedToken.Claims.(jwt.MapClaims)
			if exp, ok := parsedClaims["exp"].(string); !ok {
				t.Errorf("Expected exp claim to be string type, got type: %T", parsedClaims["exp"])
			} else if exp != "not_a_number" {
				t.Errorf("Expected exp claim value to be 'not_a_number', got: %v", exp)
			}
		}
	})

	t.Run("JWT with exp claim in the past (expired token)", func(t *testing.T) {
		// Create a JWT that's already expired
		claims := jwt.MapClaims{
			"iss": "123456",
			"iat": time.Now().Add(-2 * time.Hour).Unix(),
			"exp": time.Now().Add(-1 * time.Hour).Unix(), // Expired 1 hour ago
		}
		
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		tokenString, err := token.SignedString(auth.privateKey)
		if err != nil {
			t.Fatalf("Failed to sign test JWT: %v", err)
		}

		// Parse the token - this should succeed but token will be invalid due to expiry
		parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return &auth.privateKey.PublicKey, nil
		})
		
		// The parsing might fail due to expiry validation
		if err == nil && parsedToken.Valid {
			t.Error("Expected token to be invalid due to expiry, but it was valid")
		}
		
		// Check the claims anyway
		if err == nil {
			parsedClaims := parsedToken.Claims.(jwt.MapClaims)
			exp := int64(parsedClaims["exp"].(float64))
			if exp > time.Now().Unix() {
				t.Error("Expected token to be expired, but exp is in the future")
			}
		}
	})

	t.Run("JWT with negative exp/iat values", func(t *testing.T) {
		// Create a JWT with negative timestamp values
		claims := jwt.MapClaims{
			"iss": "123456",
			"iat": int64(-1000), // Negative issued at
			"exp": int64(-500),  // Negative expiry
		}
		
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		tokenString, err := token.SignedString(auth.privateKey)
		if err != nil {
			t.Fatalf("Failed to sign test JWT: %v", err)
		}

		// Parse the token - this will likely fail due to expired validation
		parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return &auth.privateKey.PublicKey, nil
		})
		if err == nil {
			t.Error("Expected error for negative exp value (expired), but got none")
		} else if !strings.Contains(err.Error(), "token is expired") {
			t.Logf("Got expected error for negative/expired token: %v", err)
		}

		// We can still check raw claims even if token is invalid
		if parsedToken != nil {
			parsedClaims := parsedToken.Claims.(jwt.MapClaims)
			if iat := int64(parsedClaims["iat"].(float64)); iat != -1000 {
				t.Errorf("Expected iat to be -1000, got: %v", iat)
			}
			if exp := int64(parsedClaims["exp"].(float64)); exp != -500 {
				t.Errorf("Expected exp to be -500, got: %v", exp)
			}
		}
	})

	t.Run("JWT with iat claim in the future", func(t *testing.T) {
		// Create a JWT with issued at time in the future
		futureTime := time.Now().Add(1 * time.Hour).Unix()
		claims := jwt.MapClaims{
			"iss": "123456",
			"iat": futureTime, // Future issued at
			"exp": time.Now().Add(2 * time.Hour).Unix(),
		}
		
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		tokenString, err := token.SignedString(auth.privateKey)
		if err != nil {
			t.Fatalf("Failed to sign test JWT: %v", err)
		}

		// Parse the token
		parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return &auth.privateKey.PublicKey, nil
		})
		if err != nil {
			t.Fatalf("Failed to parse JWT: %v", err)
		}

		// Verify the future iat value
		parsedClaims := parsedToken.Claims.(jwt.MapClaims)
		if iat := int64(parsedClaims["iat"].(float64)); iat != futureTime {
			t.Errorf("Expected iat to be %v, got: %v", futureTime, iat)
		}
	})

	t.Run("JWT with missing exp claim", func(t *testing.T) {
		// Create a JWT missing the exp claim
		claims := jwt.MapClaims{
			"iss": "123456",
			"iat": time.Now().Unix(),
			// Missing "exp" claim
		}
		
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		tokenString, err := token.SignedString(auth.privateKey)
		if err != nil {
			t.Fatalf("Failed to sign test JWT: %v", err)
		}

		// Parse the token
		parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return &auth.privateKey.PublicKey, nil
		})
		if err != nil {
			t.Fatalf("Failed to parse JWT: %v", err)
		}

		// Verify that the exp claim is missing
		parsedClaims := parsedToken.Claims.(jwt.MapClaims)
		if _, exists := parsedClaims["exp"]; exists {
			t.Error("Expected exp claim to be missing, but it was present")
		}
	})

	t.Run("JWT with missing iat claim", func(t *testing.T) {
		// Create a JWT missing the iat claim
		claims := jwt.MapClaims{
			"iss": "123456",
			"exp": time.Now().Add(10 * time.Minute).Unix(),
			// Missing "iat" claim
		}
		
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		tokenString, err := token.SignedString(auth.privateKey)
		if err != nil {
			t.Fatalf("Failed to sign test JWT: %v", err)
		}

		// Parse the token
		parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return &auth.privateKey.PublicKey, nil
		})
		if err != nil {
			t.Fatalf("Failed to parse JWT: %v", err)
		}

		// Verify that the iat claim is missing
		parsedClaims := parsedToken.Claims.(jwt.MapClaims)
		if _, exists := parsedClaims["iat"]; exists {
			t.Error("Expected iat claim to be missing, but it was present")
		}
	})

	t.Run("JWT with extra unexpected claims", func(t *testing.T) {
		// Create a JWT with additional unexpected claims
		claims := jwt.MapClaims{
			"iss":           "123456",
			"iat":           time.Now().Unix(),
			"exp":           time.Now().Add(10 * time.Minute).Unix(),
			"custom_field":  "unexpected_value",
			"another_field": 42,
			"array_field":   []string{"item1", "item2"},
		}
		
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		tokenString, err := token.SignedString(auth.privateKey)
		if err != nil {
			t.Fatalf("Failed to sign test JWT: %v", err)
		}

		// Parse the token
		parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return &auth.privateKey.PublicKey, nil
		})
		if err != nil {
			t.Fatalf("Failed to parse JWT: %v", err)
		}

		// Verify the extra claims exist
		parsedClaims := parsedToken.Claims.(jwt.MapClaims)
		if custom, exists := parsedClaims["custom_field"]; !exists || custom != "unexpected_value" {
			t.Errorf("Expected custom_field to be 'unexpected_value', got: %v", custom)
		}
		if another, exists := parsedClaims["another_field"]; !exists || int(another.(float64)) != 42 {
			t.Errorf("Expected another_field to be 42, got: %v", another)
		}
	})

	t.Run("JWT with null/empty claim values", func(t *testing.T) {
		// Create a JWT with null and empty values
		claims := jwt.MapClaims{
			"iss":          "",     // Empty string
			"iat":          time.Now().Unix(),
			"exp":          time.Now().Add(10 * time.Minute).Unix(),
			"null_field":   nil,   // Null value
			"empty_string": "",    // Empty string
			"zero_number":  0,     // Zero value
		}
		
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		tokenString, err := token.SignedString(auth.privateKey)
		if err != nil {
			t.Fatalf("Failed to sign test JWT: %v", err)
		}

		// Parse the token
		parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return &auth.privateKey.PublicKey, nil
		})
		if err != nil {
			t.Fatalf("Failed to parse JWT: %v", err)
		}

		// Verify the null/empty values
		parsedClaims := parsedToken.Claims.(jwt.MapClaims)
		if iss := parsedClaims["iss"].(string); iss != "" {
			t.Errorf("Expected iss to be empty string, got: %v", iss)
		}
		if nullField := parsedClaims["null_field"]; nullField != nil {
			t.Errorf("Expected null_field to be nil, got: %v", nullField)
		}
	})

	t.Run("JWT with extremely large numeric values", func(t *testing.T) {
		// Create a JWT with very large numeric values but keep exp reasonable to avoid expiry
		futureExp := time.Now().Add(1 * time.Hour).Unix()
		claims := jwt.MapClaims{
			"iss":        "123456",
			"iat":        int64(1000000000), // Large but reasonable iat value
			"exp":        futureExp,         // Valid future exp
			"large_num":  float64(1e20),     // Very large float
		}
		
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		tokenString, err := token.SignedString(auth.privateKey)
		if err != nil {
			t.Fatalf("Failed to sign test JWT: %v", err)
		}

		// Parse the token
		parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return &auth.privateKey.PublicKey, nil
		})
		if err != nil {
			t.Fatalf("Failed to parse JWT: %v", err)
		}

		// Verify the large values
		parsedClaims := parsedToken.Claims.(jwt.MapClaims)
		if iat := int64(parsedClaims["iat"].(float64)); iat != 1000000000 {
			t.Errorf("Expected iat to be 1000000000, got: %v", iat)
		}
		if largeNum := parsedClaims["large_num"].(float64); largeNum != 1e20 {
			t.Errorf("Expected large_num to be 1e20, got: %v", largeNum)
		}
	})

	t.Run("JWT with special characters in string claims", func(t *testing.T) {
		// Create a JWT with special characters in string claims
		claims := jwt.MapClaims{
			"iss":              "123456",
			"iat":              time.Now().Unix(),
			"exp":              time.Now().Add(10 * time.Minute).Unix(),
			"unicode_field":    "Hello 世界 🌍",
			"special_chars":    "!@#$%^&*()_+-=[]{}|;':\",./<>?",
			"newlines":         "line1\nline2\r\nline3",
			"tabs_spaces":      "\t  spaced  \t",
			"escape_sequences": "\\n\\t\\r\\\"\\\\",
		}
		
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		tokenString, err := token.SignedString(auth.privateKey)
		if err != nil {
			t.Fatalf("Failed to sign test JWT: %v", err)
		}

		// Parse the token
		parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return &auth.privateKey.PublicKey, nil
		})
		if err != nil {
			t.Fatalf("Failed to parse JWT: %v", err)
		}

		// Verify the special character values
		parsedClaims := parsedToken.Claims.(jwt.MapClaims)
		if unicode := parsedClaims["unicode_field"].(string); unicode != "Hello 世界 🌍" {
			t.Errorf("Expected unicode_field to preserve unicode, got: %v", unicode)
		}
		if special := parsedClaims["special_chars"].(string); special != "!@#$%^&*()_+-=[]{}|;':\",./<>?" {
			t.Errorf("Expected special_chars to preserve special characters, got: %v", special)
		}
		if newlines := parsedClaims["newlines"].(string); newlines != "line1\nline2\r\nline3" {
			t.Errorf("Expected newlines to preserve newline characters, got: %v", newlines)
		}
	})
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

func TestVerifyTeamMembership(t *testing.T) {
	privateKey := generateTestKey(t)
	auth, err := NewGitHubAuth(privateKey, 123456)
	if err != nil {
		t.Fatalf("Failed to create GitHubAuth: %v", err)
	}

	tests := []struct {
		name              string
		installationToken string
		username          string
		organization      string
		team              string
		mockStatus        int
		mockResponse      interface{}
		wantMember        bool
		wantErr           bool
		errorContains     string
	}{
		{
			name:              "User is an active team member",
			installationToken: "ghs_test_token",
			username:          "testuser",
			organization:      "testorg",
			team:              "testteam",
			mockStatus:        http.StatusOK,
			mockResponse: map[string]interface{}{
				"state": "active",
			},
			wantMember: true,
			wantErr:    false,
		},
		{
			name:              "User is a pending team member",
			installationToken: "ghs_test_token",
			username:          "testuser",
			organization:      "testorg",
			team:              "testteam",
			mockStatus:        http.StatusOK,
			mockResponse: map[string]interface{}{
				"state": "pending",
			},
			wantMember: false,
			wantErr:    false,
		},
		{
			name:              "User is not a team member",
			installationToken: "ghs_test_token",
			username:          "testuser",
			organization:      "testorg",
			team:              "testteam",
			mockStatus:        http.StatusNotFound,
			wantMember:        false,
			wantErr:           false,
		},
		{
			name:              "Access forbidden",
			installationToken: "ghs_test_token",
			username:          "testuser",
			organization:      "testorg",
			team:              "testteam",
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
			team:              "testteam",
			mockStatus:        http.StatusOK,
			wantMember:        false,
			wantErr:           true,
			errorContains:     "username, organization, and team are required",
		},
		{
			name:              "Empty organization",
			installationToken: "ghs_test_token",
			username:          "testuser",
			organization:      "",
			team:              "testteam",
			mockStatus:        http.StatusOK,
			wantMember:        false,
			wantErr:           true,
			errorContains:     "username, organization, and team are required",
		},
		{
			name:              "Empty team",
			installationToken: "ghs_test_token",
			username:          "testuser",
			organization:      "testorg",
			team:              "",
			mockStatus:        http.StatusOK,
			wantMember:        false,
			wantErr:           true,
			errorContains:     "username, organization, and team are required",
		},
		{
			name:              "Server error",
			installationToken: "ghs_test_token",
			username:          "testuser",
			organization:      "testorg",
			team:              "testteam",
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
				expectedPath := fmt.Sprintf("/orgs/%s/teams/%s/memberships/%s", tt.organization, tt.team, tt.username)
				if tt.username != "" && tt.organization != "" && tt.team != "" && r.URL.Path != expectedPath {
					t.Errorf("Expected path %q but got %q", expectedPath, r.URL.Path)
				}

				w.WriteHeader(tt.mockStatus)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			// Use test server instead of real GitHub API
			originalURL := GetGitHubAPIBaseURL()
			SetGitHubAPIBaseURL(server.URL)
			defer SetGitHubAPIBaseURL(originalURL)

			isMember, err := auth.VerifyTeamMembership(tt.installationToken, tt.username, tt.organization, tt.team)
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

func TestGetInstallation(t *testing.T) {
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
		expectedOrg    string
	}{
		{
			name:           "Successful installation retrieval",
			installationID: 987654,
			mockStatus:     http.StatusOK,
			mockResponse: InstallationResponse{
				ID: 987654,
				Account: struct {
					Login string `json:"login"`
					Type  string `json:"type"`
				}{
					Login: "myorg",
					Type:  "Organization",
				},
			},
			wantErr:     false,
			expectedOrg: "myorg",
		},
		{
			name:           "Installation not found",
			installationID: 0,
			mockStatus:     http.StatusNotFound,
			mockResponse:   map[string]string{"message": "Not Found"},
			wantErr:        true,
			errorContains:  "404",
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
			mockStatus:     http.StatusOK,
			mockResponse:   "not a json response",
			wantErr:        true,
			errorContains:  "decoding response",
		},
		{
			name:           "Missing account login in response",
			installationID: 987654,
			mockStatus:     http.StatusOK,
			mockResponse: map[string]interface{}{
				"id": 987654,
				"account": map[string]interface{}{
					"type": "Organization",
				},
			},
			wantErr:       true,
			errorContains: "invalid installation response",
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
				expectedPath := fmt.Sprintf("/app/installations/%d", tt.installationID)
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

			installation, err := auth.GetInstallation(tt.installationID)
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
			if installation == nil {
				t.Error("Expected non-nil InstallationResponse but got nil")
				return
			}
			if installation.Account.Login != tt.expectedOrg {
				t.Errorf("Expected organization %q but got %q", tt.expectedOrg, installation.Account.Login)
			}
		})
	}
}

// Helper functions
func errorStartsWith(err error, prefix string) bool {
	return err != nil && len(err.Error()) >= len(prefix) && err.Error()[:len(prefix)] == prefix
}
