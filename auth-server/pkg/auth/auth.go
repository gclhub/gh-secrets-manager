package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type GitHubAuth struct {
	privateKey *rsa.PrivateKey
	appID      int64
}

type TokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func NewGitHubAuth(privateKeyPEM []byte, appID int64) (*GitHubAuth, error) {
	log.Printf("Initializing GitHub auth for app-id=%d", appID)

	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		log.Printf("Failed to decode PEM block for app-id=%d", appID)
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	log.Printf("Successfully decoded PEM block for app-id=%d", appID)

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		log.Printf("Failed to parse private key for app-id=%d: %v", appID, err)
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	log.Printf("Successfully parsed private key for app-id=%d", appID)

	return &GitHubAuth{
		privateKey: privateKey,
		appID:      appID,
	}, nil
}

func (gh *GitHubAuth) GenerateJWT() (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(10 * time.Minute)),
		Issuer:    fmt.Sprintf("%d", gh.appID),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	log.Printf("Generating JWT for app-id=%d, expires=%s", gh.appID, claims.ExpiresAt.Time)

	signedToken, err := token.SignedString(gh.privateKey)
	if err != nil {
		log.Printf("Failed to sign JWT for app-id=%d: %v", gh.appID, err)
		return "", fmt.Errorf("signing token: %w", err)
	}

	log.Printf("Successfully generated JWT for app-id=%d", gh.appID)
	return signedToken, nil
}

func (gh *GitHubAuth) GetInstallationToken(installationID int64) (*TokenResponse, error) {
	jwt, err := gh.GenerateJWT()
	if err != nil {
		return nil, fmt.Errorf("generating JWT: %w", err)
	}

	url := fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", installationID)
	log.Printf("Requesting installation token from GitHub API: %s", url)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		log.Printf("Failed to create GitHub API request: %v", err)
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", fmt.Sprintf("GitHubApp/%d", gh.appID))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to make GitHub API request: %v", err)
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read GitHub API response: %v", err)
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		log.Printf("GitHub API error: status=%d body=%s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expires_at"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		log.Printf("Failed to parse GitHub API response: %v", err)
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	log.Printf("Successfully obtained installation token from GitHub API, expires=%s", tokenResp.ExpiresAt)
	return &TokenResponse{
		Token:     tokenResp.Token,
		ExpiresAt: tokenResp.ExpiresAt,
	}, nil
}
