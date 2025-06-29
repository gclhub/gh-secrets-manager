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

var Verbose = false

type GitHubAuth struct {
	privateKey *rsa.PrivateKey
	appID      int64
}

type TokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type InstallationResponse struct {
	ID      int64 `json:"id"`
	Account struct {
		Login string `json:"login"`
		Type  string `json:"type"`
	} `json:"account"`
}

func NewGitHubAuth(privateKeyPEM []byte, appID int64) (*GitHubAuth, error) {
	if Verbose {
		log.Printf("Initializing GitHub auth for app-id=%d", appID)
	}

	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		if Verbose {
			log.Printf("Failed to decode PEM block for app-id=%d", appID)
		}
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	if Verbose {
		log.Printf("Successfully decoded PEM block for app-id=%d", appID)
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		if Verbose {
			log.Printf("Failed to parse private key for app-id=%d: %v", appID, err)
		}
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	if Verbose {
		log.Printf("Successfully parsed private key for app-id=%d", appID)
	}

	return &GitHubAuth{
		privateKey: privateKey,
		appID:      appID,
	}, nil
}

func (gh *GitHubAuth) GenerateJWT() (string, error) {
	if gh.privateKey == nil {
		return "", fmt.Errorf("private key is nil")
	}
	now := time.Now()
	claims := jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(10 * time.Minute)),
		Issuer:    fmt.Sprintf("%d", gh.appID),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	if Verbose {
		log.Printf("Generating JWT for app-id=%d, expires=%s", gh.appID, claims.ExpiresAt.Time)
	}

	signedToken, err := token.SignedString(gh.privateKey)
	if err != nil {
		if Verbose {
			log.Printf("Failed to sign JWT for app-id=%d: %v", gh.appID, err)
		}
		return "", fmt.Errorf("signing token: %w", err)
	}

	if Verbose {
		log.Printf("Successfully generated JWT for app-id=%d", gh.appID)
	}
	return signedToken, nil
}

func (gh *GitHubAuth) GetInstallationToken(installationID int64) (*TokenResponse, error) {
	jwt, err := gh.GenerateJWT()
	if err != nil {
		return nil, fmt.Errorf("generating JWT: %w", err)
	}

	url := fmt.Sprintf("%s/app/installations/%d/access_tokens", GetGitHubAPIBaseURL(), installationID)
	if Verbose {
		log.Printf("Requesting installation token from GitHub API: %s", url)
	}

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		if Verbose {
			log.Printf("Failed to create GitHub API request: %v", err)
		}
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", fmt.Sprintf("GitHubApp/%d", gh.appID))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if Verbose {
			log.Printf("Failed to make GitHub API request: %v", err)
		}
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if Verbose {
			log.Printf("Failed to read GitHub API response: %v", err)
		}
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		if Verbose {
			log.Printf("GitHub API error: status=%d body=%s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expires_at"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		if Verbose {
			log.Printf("Failed to parse GitHub API response: %v", err)
		}
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	// Validate required fields
	if tokenResp.Token == "" || tokenResp.ExpiresAt.IsZero() {
		if Verbose {
			log.Printf("Invalid token response: missing required fields")
		}
		return nil, fmt.Errorf("invalid token response: missing required fields")
	}

	if Verbose {
		log.Printf("Successfully obtained installation token from GitHub API, expires=%s", tokenResp.ExpiresAt)
	}
	return &TokenResponse{
		Token:     tokenResp.Token,
		ExpiresAt: tokenResp.ExpiresAt,
	}, nil
}

// GetInstallation retrieves information about a GitHub App installation
func (gh *GitHubAuth) GetInstallation(installationID int64) (*InstallationResponse, error) {
	jwt, err := gh.GenerateJWT()
	if err != nil {
		return nil, fmt.Errorf("generating JWT: %w", err)
	}

	url := fmt.Sprintf("%s/app/installations/%d", GetGitHubAPIBaseURL(), installationID)
	if Verbose {
		log.Printf("Requesting installation details from GitHub API: %s", url)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		if Verbose {
			log.Printf("Failed to create GitHub API request: %v", err)
		}
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", fmt.Sprintf("GitHubApp/%d", gh.appID))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if Verbose {
			log.Printf("Failed to make GitHub API request: %v", err)
		}
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if Verbose {
			log.Printf("Failed to read GitHub API response: %v", err)
		}
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if Verbose {
			log.Printf("GitHub API error: status=%d body=%s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(body))
	}

	var installation InstallationResponse
	if err := json.Unmarshal(body, &installation); err != nil {
		if Verbose {
			log.Printf("Failed to parse GitHub API response: %v", err)
		}
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	// Validate required fields
	if installation.Account.Login == "" {
		if Verbose {
			log.Printf("Invalid installation response: missing account login")
		}
		return nil, fmt.Errorf("invalid installation response: missing account login")
	}

	if Verbose {
		log.Printf("Successfully obtained installation details: org=%s type=%s", installation.Account.Login, installation.Account.Type)
	}
	return &installation, nil
}

// VerifyTeamMembership checks if a user belongs to the specified team within an organization
func (gh *GitHubAuth) VerifyTeamMembership(installationToken, username, organization, team string) (bool, error) {
	if username == "" || organization == "" || team == "" {
		return false, fmt.Errorf("username, organization, and team are required")
	}

	url := fmt.Sprintf("%s/orgs/%s/teams/%s/memberships/%s", GetGitHubAPIBaseURL(), organization, team, username)
	if Verbose {
		log.Printf("Checking team membership: %s", url)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		if Verbose {
			log.Printf("Failed to create team membership check request: %v", err)
		}
		return false, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+installationToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", fmt.Sprintf("GitHubApp/%d", gh.appID))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if Verbose {
			log.Printf("Failed to make team membership check request: %v", err)
		}
		return false, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// User is a team member - verify the membership details
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			if Verbose {
				log.Printf("Failed to read team membership response: %v", err)
			}
			return false, fmt.Errorf("reading response: %w", err)
		}
		
		var membership struct {
			State string `json:"state"`
		}
		if err := json.Unmarshal(body, &membership); err != nil {
			if Verbose {
				log.Printf("Failed to parse team membership response: %v", err)
			}
			return false, fmt.Errorf("parsing response: %w", err)
		}

		// Check if membership is active
		isActive := membership.State == "active"
		if Verbose {
			log.Printf("User %s team membership in %s/%s: state=%s, active=%v", username, organization, team, membership.State, isActive)
		}
		return isActive, nil
	case http.StatusNotFound:
		// User is not a team member
		if Verbose {
			log.Printf("User %s is not a member of team %s in organization %s", username, team, organization)
		}
		return false, nil
	case http.StatusForbidden:
		// Requester does not have permission to check team membership
		if Verbose {
			log.Printf("Access forbidden when checking team membership for %s in %s/%s", username, organization, team)
		}
		return false, fmt.Errorf("access forbidden: unable to check team membership")
	default:
		body, _ := io.ReadAll(resp.Body)
		if Verbose {
			log.Printf("Unexpected response from GitHub API: status=%d body=%s", resp.StatusCode, string(body))
		}
		return false, fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(body))
	}
}
