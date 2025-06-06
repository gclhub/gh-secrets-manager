package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"gh-secrets-manager/pkg/config"

	"github.com/cli/go-gh"
	"github.com/google/go-github/v45/github"
	"golang.org/x/crypto/nacl/box"
)

// Verbose controls the logging verbosity across the API package
var Verbose bool

type AuthMethod int

const (
	AuthMethodPAT AuthMethod = iota
	AuthMethodGitHubApp
)

type ClientOptions struct {
	AuthMethod     AuthMethod
	AppID          int64
	InstallationID int64
	AuthServer     string
	Username       string
	Organization   string
	Team           string
}

type authResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Client struct {
	github    *github.Client
	ctx       context.Context
	opts      *ClientOptions
	authToken string
	expiresAt time.Time
}

func GetCurrentUsername() (string, error) {
	// Use gh CLI to get current username
	client, err := gh.RESTClient(nil)
	if err != nil {
		return "", fmt.Errorf("failed to create GitHub client: %w", err)
	}

	var user struct {
		Login string `json:"login"`
	}
	
	err = client.Get("user", &user)
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}

	return user.Login, nil
}

func NewClient() (*Client, error) {
	// Try to load GitHub App config first
	if Verbose {
		log.Printf("Loading configuration...")
	}
	cfg, err := config.Load()
	if err != nil {
		if Verbose {
			log.Printf("Failed to load config: %v, falling back to PAT auth", err)
		}
		return NewClientWithOptions(&ClientOptions{AuthMethod: AuthMethodPAT})
	}

	// If GitHub App is configured, use it as default
	if cfg.IsGitHubAppConfigured() {
		if Verbose {
			log.Printf("Using GitHub App authentication (app-id=%d, installation-id=%d)", cfg.AppID, cfg.InstallationID)
		}
		
		// Get current username for organization verification if using GitHub App auth
		username, err := GetCurrentUsername()
		if err != nil {
			if Verbose {
				log.Printf("Warning: Failed to get current username for organization verification: %v", err)
			}
			// Continue without username - auth server may not require organization verification
		}
		
		return NewClientWithOptions(&ClientOptions{
			AuthMethod:     AuthMethodGitHubApp,
			AppID:          cfg.AppID,
			InstallationID: cfg.InstallationID,
			AuthServer:     cfg.AuthServer,
			Username:       username,
			Organization:   cfg.Organization,
			Team:           cfg.Team,
		})
	}

	// Only fall back to PAT if GitHub App is not configured
	if Verbose {
		log.Printf("GitHub App configuration not found, falling back to PAT authentication")
	}
	return NewClientWithOptions(&ClientOptions{AuthMethod: AuthMethodPAT})
}

func NewClientWithOptions(opts *ClientOptions) (*Client, error) {
	if opts == nil {
		if Verbose {
			log.Printf("No options provided, using default PAT auth")
		}
		return newPATClient()
	}

	switch opts.AuthMethod {
	case AuthMethodPAT:
		if Verbose {
			log.Printf("Using PAT authentication")
		}
		return newPATClient()

	case AuthMethodGitHubApp:
		if Verbose {
			log.Printf("Initializing GitHub App client (auth-server=%s, app-id=%d, installation-id=%d, username=%s, org=%s, team=%s)",
				opts.AuthServer, opts.AppID, opts.InstallationID, opts.Username, opts.Organization, opts.Team)
		}
		client := &Client{
			ctx:    context.Background(),
			github: github.NewClient(&http.Client{}),
			opts:   opts,
		}

		// Get initial token
		if err := client.refreshToken(); err != nil {
			log.Printf("Failed to get initial GitHub App token: %v", err)
			return nil, fmt.Errorf("failed to get initial token: %w", err)
		}
		log.Printf("Successfully obtained GitHub App token, expires at %s", client.expiresAt)

		// Update GitHub client with the token
		client.github = github.NewClient(&http.Client{
			Transport: &authorizedTransport{
				token: client.authToken,
			},
		})

		return client, nil

	default:
		return nil, fmt.Errorf("unsupported authentication method")
	}
}

type restTransport struct {
	client any // gh.RESTClient interface
}

func (t *restTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// gh.RESTClient implements a Do method that handles auth
	if client, ok := t.client.(interface {
		Do(*http.Request) (*http.Response, error)
	}); ok {
		return client.Do(req)
	}
	return nil, fmt.Errorf("invalid REST client type")
}

func newPATClient() (*Client, error) {
	// Use gh CLI's built-in REST client which handles auth
	restClient, err := gh.RESTClient(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub client: %w", err)
	}

	// Create a transport that uses the REST client directly
	httpClient := &http.Client{
		Transport: &restTransport{
			client: restClient,
		},
	}

	return &Client{
		github: github.NewClient(httpClient),
		ctx:    context.Background(),
		opts:   &ClientOptions{AuthMethod: AuthMethodPAT},
	}, nil
}

func (c *Client) refreshToken() error {
	if c.opts.AuthServer == "" {
		if Verbose {
			log.Printf("No auth server URL provided")
		}
		return fmt.Errorf("auth server URL is required")
	}

	// Clean up auth server URL by trimming trailing slash
	authServer := strings.TrimRight(c.opts.AuthServer, "/")
	tokenURL := fmt.Sprintf("%s/token", authServer)
	if Verbose {
		log.Printf("Requesting token from auth server: %s", tokenURL)
	}

	req, err := http.NewRequest("POST", tokenURL, nil)
	if err != nil {
		if Verbose {
			log.Printf("Failed to create auth request: %v", err)
		}
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	q := req.URL.Query()
	q.Add("app-id", fmt.Sprintf("%d", c.opts.AppID))
	q.Add("installation-id", fmt.Sprintf("%d", c.opts.InstallationID))
	
	// Add username if provided for team verification
	if c.opts.Username != "" {
		q.Add("username", c.opts.Username)
		if Verbose {
			log.Printf("Adding username to auth request: %s", c.opts.Username)
		}
	}
	
	// Add organization if provided for team verification
	if c.opts.Organization != "" {
		q.Add("org", c.opts.Organization)
		if Verbose {
			log.Printf("Adding organization to auth request: %s", c.opts.Organization)
		}
	}
	
	// Add team if provided for team verification
	if c.opts.Team != "" {
		q.Add("team", c.opts.Team)
		if Verbose {
			log.Printf("Adding team to auth request: %s", c.opts.Team)
		}
	}
	
	req.URL.RawQuery = q.Encode()

	if Verbose {
		log.Printf("Making request to auth server with URL: %s", req.URL.String())
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if Verbose {
			log.Printf("Failed to get token from auth server: %v", err)
		}
		return fmt.Errorf("failed to get token from auth server: %w", err)
	}
	defer resp.Body.Close()

	if Verbose {
		log.Printf("Auth server response status: %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		if Verbose {
			log.Printf("Auth server error response: %s", string(bodyBytes))
		}
		return fmt.Errorf("auth server returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var authResp authResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		if Verbose {
			log.Printf("Failed to decode auth response: %v", err)
		}
		return fmt.Errorf("failed to decode auth response: %w", err)
	}

	if Verbose {
		log.Printf("Successfully obtained new token, expires at: %s", authResp.ExpiresAt)
	}
	c.authToken = authResp.Token
	c.expiresAt = authResp.ExpiresAt

	// Update the GitHub client with the new token
	c.github = github.NewClient(&http.Client{
		Transport: &authorizedTransport{
			token: c.authToken,
		},
	})

	return nil
}

type authorizedTransport struct {
	token string
}

func (t *authorizedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	// Add User-Agent as required by GitHub API
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "gh-secrets-manager")
	}
	return http.DefaultTransport.RoundTrip(req)
}

func (c *Client) ensureValidToken() error {
	if c.opts.AuthMethod != AuthMethodGitHubApp {
		return nil
	}

	// Refresh token if it's expired or will expire in the next minute
	if time.Now().Add(time.Minute).After(c.expiresAt) {
		if Verbose {
			log.Printf("Token expired or will expire soon (expires at: %s), refreshing", c.expiresAt)
		}
		if err := c.refreshToken(); err != nil {
			if Verbose {
				log.Printf("Failed to refresh token: %v", err)
			}
			return fmt.Errorf("failed to refresh token: %w", err)
		}
		if Verbose {
			log.Printf("Successfully refreshed token, new expiry: %s", c.expiresAt)
		}
	}

	return nil
}

// Variable represents a GitHub Actions variable
type Variable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Secrets methods
func (c *Client) ListOrgSecrets(org string) ([]*github.Secret, error) {
	if err := c.ensureValidToken(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("orgs/%s/actions/secrets", org)
	req, err := c.github.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var response struct {
		Secrets []*github.Secret `json:"secrets"`
	}
	_, err = c.github.Do(c.ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list organization secrets: %w", err)
	}
	return response.Secrets, nil
}

func (c *Client) ListRepoSecrets(owner, repo string) ([]*github.Secret, error) {
	if err := c.ensureValidToken(); err != nil {
		return nil, err
	}

	secrets, _, err := c.github.Actions.ListRepoSecrets(c.ctx, owner, repo, &github.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list repository secrets: %w", err)
	}
	return secrets.Secrets, nil
}

func (c *Client) CreateOrUpdateOrgSecret(org string, secret *github.EncryptedSecret) error {
	if err := c.ensureValidToken(); err != nil {
		return err
	}

	encryption, err := c.GetOrgPublicKey(org)
	if err != nil {
		return err
	}

	encryptedSecret, err := encryption.CreateEncryptedSecret(secret.Name, secret.EncryptedValue)
	if err != nil {
		return err
	}

	// Custom implementation that uses github.Client's underlying HTTP client
	// instead of using github.Actions.CreateOrUpdateOrgSecret
	url := fmt.Sprintf("orgs/%s/actions/secrets/%s", org, encryptedSecret.Name)
	req := struct {
		EncryptedValue string `json:"encrypted_value"`
		KeyID          string `json:"key_id"`
	}{
		EncryptedValue: encryptedSecret.EncryptedValue,
		KeyID:          encryptedSecret.KeyID,
	}

	httpReq, err := c.github.NewRequest("PUT", url, req)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = c.github.Do(c.ctx, httpReq, nil)
	if err != nil {
		return fmt.Errorf("failed to create/update organization secret: %w", err)
	}

	return nil
}

func (c *Client) CreateOrUpdateRepoSecret(owner, repo string, secret *github.EncryptedSecret) error {
	if err := c.ensureValidToken(); err != nil {
		return err
	}

	encryption, err := c.GetRepoPublicKey(owner, repo)
	if err != nil {
		return err
	}

	encryptedSecret, err := encryption.CreateEncryptedSecret(secret.Name, secret.EncryptedValue)
	if err != nil {
		return err
	}

	// Custom implementation that uses github.Client's underlying HTTP client
	// instead of using github.Actions.CreateOrUpdateRepoSecret
	url := fmt.Sprintf("repos/%s/%s/actions/secrets/%s", owner, repo, encryptedSecret.Name)
	req := struct {
		EncryptedValue string `json:"encrypted_value"`
		KeyID          string `json:"key_id"`
	}{
		EncryptedValue: encryptedSecret.EncryptedValue,
		KeyID:          encryptedSecret.KeyID,
	}

	httpReq, err := c.github.NewRequest("PUT", url, req)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = c.github.Do(c.ctx, httpReq, nil)
	if err != nil {
		return fmt.Errorf("failed to create/update repository secret: %w", err)
	}

	return nil
}

func (c *Client) DeleteOrgSecret(org, secretName string) error {
	if err := c.ensureValidToken(); err != nil {
		return err
	}

	// Custom implementation that uses github.Client's underlying HTTP client
	// instead of using github.Actions.DeleteOrgSecret
	url := fmt.Sprintf("orgs/%s/actions/secrets/%s", org, secretName)
	req, err := c.github.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = c.github.Do(c.ctx, req, nil)
	if err != nil {
		return fmt.Errorf("failed to delete organization secret: %w", err)
	}

	return nil
}

func (c *Client) DeleteRepoSecret(owner, repo, secretName string) error {
	if err := c.ensureValidToken(); err != nil {
		return err
	}

	// Custom implementation that uses github.Client's underlying HTTP client
	// instead of using github.Actions.DeleteRepoSecret
	url := fmt.Sprintf("repos/%s/%s/actions/secrets/%s", owner, repo, secretName)
	req, err := c.github.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = c.github.Do(c.ctx, req, nil)
	if err != nil {
		return fmt.Errorf("failed to delete repository secret: %w", err)
	}

	return nil
}

// Variables methods - implemented using custom API calls since the go-github library
// doesn't support variables yet
func (c *Client) ListOrgVariables(org string) ([]*Variable, error) {
	if err := c.ensureValidToken(); err != nil {
		return nil, err
	}

	req, err := c.github.NewRequest("GET", fmt.Sprintf("orgs/%s/actions/variables", org), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var response struct {
		Variables []*Variable `json:"variables"`
	}
	_, err = c.github.Do(c.ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list organization variables: %w", err)
	}

	return response.Variables, nil
}

func (c *Client) ListRepoVariables(owner, repo string) ([]*Variable, error) {
	if err := c.ensureValidToken(); err != nil {
		return nil, err
	}

	req, err := c.github.NewRequest("GET", fmt.Sprintf("repos/%s/%s/actions/variables", owner, repo), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var response struct {
		Variables []*Variable `json:"variables"`
	}
	_, err = c.github.Do(c.ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list repository variables: %w", err)
	}

	return response.Variables, nil
}

func (c *Client) CreateOrUpdateOrgVariable(org string, variable *Variable) error {
	if err := c.ensureValidToken(); err != nil {
		return err
	}

	url := fmt.Sprintf("orgs/%s/actions/variables/%s", org, variable.Name)
	req, err := c.github.NewRequest("PATCH", url, variable)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = c.github.Do(c.ctx, req, nil)
	if err != nil {
		return fmt.Errorf("failed to create/update organization variable: %w", err)
	}
	return nil
}

func (c *Client) CreateOrUpdateRepoVariable(owner, repo string, variable *Variable) error {
	if err := c.ensureValidToken(); err != nil {
		return err
	}

	url := fmt.Sprintf("repos/%s/%s/actions/variables/%s", owner, repo, variable.Name)
	req, err := c.github.NewRequest("PATCH", url, variable)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = c.github.Do(c.ctx, req, nil)
	if err != nil {
		return fmt.Errorf("failed to create/update repository variable: %w", err)
	}
	return nil
}

func (c *Client) DeleteOrgVariable(org, variableName string) error {
	if err := c.ensureValidToken(); err != nil {
		return err
	}

	url := fmt.Sprintf("orgs/%s/actions/variables/%s", org, variableName)
	req, err := c.github.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = c.github.Do(c.ctx, req, nil)
	if err != nil {
		return fmt.Errorf("failed to delete organization variable: %w", err)
	}
	return nil
}

func (c *Client) DeleteRepoVariable(owner, repo, variableName string) error {
	if err := c.ensureValidToken(); err != nil {
		return err
	}

	url := fmt.Sprintf("repos/%s/%s/actions/variables/%s", owner, repo, variableName)
	req, err := c.github.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = c.github.Do(c.ctx, req, nil)
	if err != nil {
		return fmt.Errorf("failed to delete repository variable: %w", err)
	}
	return nil
}

// Dependabot secrets methods
func (c *Client) ListOrgDependabotSecrets(org string) ([]*github.Secret, error) {
	if err := c.ensureValidToken(); err != nil {
		return nil, err
	}

	req, err := c.github.NewRequest("GET", fmt.Sprintf("orgs/%s/dependabot/secrets", org), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var response struct {
		Secrets []*github.Secret `json:"secrets"`
	}
	_, err = c.github.Do(c.ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list organization Dependabot secrets: %w", err)
	}

	return response.Secrets, nil
}

func (c *Client) ListRepoDependabotSecrets(owner, repo string) ([]*github.Secret, error) {
	if err := c.ensureValidToken(); err != nil {
		return nil, err
	}

	req, err := c.github.NewRequest("GET", fmt.Sprintf("repos/%s/%s/dependabot/secrets", owner, repo), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var response struct {
		Secrets []*github.Secret `json:"secrets"`
	}
	_, err = c.github.Do(c.ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list repository Dependabot secrets: %w", err)
	}

	return response.Secrets, nil
}

func (c *Client) CreateOrUpdateOrgDependabotSecret(org string, secret *github.EncryptedSecret) error {
	if err := c.ensureValidToken(); err != nil {
		return err
	}

	encryption, err := c.GetOrgDependabotPublicKey(org)
	if err != nil {
		return err
	}

	encryptedSecret, err := encryption.CreateEncryptedSecret(secret.Name, secret.EncryptedValue)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("orgs/%s/dependabot/secrets/%s", org, secret.Name)
	req := struct {
		EncryptedValue string `json:"encrypted_value"`
		KeyID          string `json:"key_id"`
	}{
		EncryptedValue: encryptedSecret.EncryptedValue,
		KeyID:          encryptedSecret.KeyID,
	}

	httpReq, err := c.github.NewRequest("PUT", url, req)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = c.github.Do(c.ctx, httpReq, nil)
	if err != nil {
		return fmt.Errorf("failed to create/update organization Dependabot secret: %w", err)
	}
	return nil
}

func (c *Client) CreateOrUpdateRepoDependabotSecret(owner, repo string, secret *github.EncryptedSecret) error {
	if err := c.ensureValidToken(); err != nil {
		return err
	}

	encryption, err := c.GetRepoDependabotPublicKey(owner, repo)
	if err != nil {
		return err
	}

	encryptedSecret, err := encryption.CreateEncryptedSecret(secret.Name, secret.EncryptedValue)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("repos/%s/%s/dependabot/secrets/%s", owner, repo, secret.Name)
	req, err := c.github.NewRequest("PUT", url, encryptedSecret)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = c.github.Do(c.ctx, req, nil)
	if err != nil {
		return fmt.Errorf("failed to create/update repository Dependabot secret: %w", err)
	}
	return nil
}

func (c *Client) DeleteOrgDependabotSecret(org, secretName string) error {
	if err := c.ensureValidToken(); err != nil {
		return err
	}

	url := fmt.Sprintf("orgs/%s/dependabot/secrets/%s", org, secretName)
	req, err := c.github.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = c.github.Do(c.ctx, req, nil)
	if err != nil {
		return fmt.Errorf("failed to delete organization Dependabot secret: %w", err)
	}
	return nil
}

func (c *Client) DeleteRepoDependabotSecret(owner, repo, secretName string) error {
	if err := c.ensureValidToken(); err != nil {
		return err
	}

	url := fmt.Sprintf("repos/%s/%s/dependabot/secrets/%s", owner, repo, secretName)
	req, err := c.github.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = c.github.Do(c.ctx, req, nil)
	if err != nil {
		return fmt.Errorf("failed to delete repository Dependabot secret: %w", err)
	}
	return nil
}

// Environment secrets methods
func (c *Client) GetEnvPublicKey(owner, repo, environment string) (*github.PublicKey, error) {
	if err := c.ensureValidToken(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("repos/%s/%s/environments/%s/secrets/public-key", owner, repo, environment)
	req, err := c.github.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	key := &github.PublicKey{}
	_, err = c.github.Do(c.ctx, req, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get environment public key: %w", err)
	}
	return key, nil
}

func (c *Client) ListEnvSecrets(owner, repo, environment string) ([]*github.Secret, error) {
	if err := c.ensureValidToken(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("repos/%s/%s/environments/%s/secrets", owner, repo, environment)
	req, err := c.github.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var secrets struct {
		Secrets []*github.Secret `json:"secrets"`
	}
	_, err = c.github.Do(c.ctx, req, &secrets)
	if err != nil {
		return nil, fmt.Errorf("failed to list environment secrets: %w", err)
	}
	return secrets.Secrets, nil
}

func (c *Client) CreateOrUpdateEnvSecret(owner, repo, environment string, secret *github.EncryptedSecret) error {
	if err := c.ensureValidToken(); err != nil {
		return err
	}

	key, err := c.GetEnvPublicKey(owner, repo, environment)
	if err != nil {
		return err
	}

	decodedKey, err := base64.StdEncoding.DecodeString(*key.Key)
	if err != nil {
		return fmt.Errorf("failed to decode public key: %w", err)
	}

	var publicKey [32]byte
	copy(publicKey[:], decodedKey)

	encryptedBytes, err := box.SealAnonymous(nil, []byte(secret.EncryptedValue), &publicKey, nil)
	if err != nil {
		return fmt.Errorf("failed to encrypt secret: %w", err)
	}

	req := struct {
		EncryptedValue string `json:"encrypted_value"`
		KeyID          string `json:"key_id"`
	}{
		EncryptedValue: base64.StdEncoding.EncodeToString(encryptedBytes),
		KeyID:          *key.KeyID,
	}

	url := fmt.Sprintf("repos/%s/%s/environments/%s/secrets/%s", owner, repo, environment, secret.Name)
	httpReq, err := c.github.NewRequest("PUT", url, req)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = c.github.Do(c.ctx, httpReq, nil)
	if err != nil {
		return fmt.Errorf("failed to create/update environment secret: %w", err)
	}
	return nil
}

func (c *Client) DeleteEnvSecret(owner, repo, environment, name string) error {
	if err := c.ensureValidToken(); err != nil {
		return err
	}

	url := fmt.Sprintf("repos/%s/%s/environments/%s/secrets/%s", owner, repo, environment, name)
	req, err := c.github.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = c.github.Do(c.ctx, req, nil)
	if err != nil {
		return fmt.Errorf("failed to delete environment secret: %w", err)
	}
	return nil
}

// Environment variables methods
func (c *Client) ListEnvironmentVariables(owner, repo, environment string) ([]*Variable, error) {
	if err := c.ensureValidToken(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("repos/%s/%s/environments/%s/variables", owner, repo, environment)
	req, err := c.github.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var response struct {
		Variables []*Variable `json:"variables"`
	}
	_, err = c.github.Do(c.ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list environment variables: %w", err)
	}

	return response.Variables, nil
}

func (c *Client) CreateOrUpdateEnvironmentVariable(owner, repo, environment string, variable *Variable) error {
	if err := c.ensureValidToken(); err != nil {
		return err
	}

	url := fmt.Sprintf("repos/%s/%s/environments/%s/variables/%s", owner, repo, environment, variable.Name)
	req, err := c.github.NewRequest("PUT", url, variable)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = c.github.Do(c.ctx, req, nil)
	if err != nil {
		return fmt.Errorf("failed to create/update environment variable: %w", err)
	}
	return nil
}

func (c *Client) DeleteEnvironmentVariable(owner, repo, environment, name string) error {
	if err := c.ensureValidToken(); err != nil {
		return err
	}

	url := fmt.Sprintf("repos/%s/%s/environments/%s/variables/%s", owner, repo, environment, name)
	req, err := c.github.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = c.github.Do(c.ctx, req, nil)
	if err != nil {
		return fmt.Errorf("failed to delete environment variable: %w", err)
	}
	return nil
}

// Wrappers for environment variable methods to match test expectations
func (c *Client) ListEnvVariables(owner, repo, environment string) ([]*Variable, error) {
	return c.ListEnvironmentVariables(owner, repo, environment)
}

func (c *Client) CreateOrUpdateEnvVariable(owner, repo, environment string, variable *Variable) error {
	return c.CreateOrUpdateEnvironmentVariable(owner, repo, environment, variable)
}

func (c *Client) DeleteEnvVariable(owner, repo, environment, name string) error {
	return c.DeleteEnvironmentVariable(owner, repo, environment, name)
}
