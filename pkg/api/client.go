package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cli/go-gh"
	"github.com/google/go-github/v45/github"
	"golang.org/x/crypto/nacl/box"
)

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

func NewClient() (*Client, error) {
	return NewClientWithOptions(nil)
}

func NewClientWithOptions(opts *ClientOptions) (*Client, error) {
	if opts == nil {
		// Default to PAT authentication using gh CLI
		_, err := gh.RESTClient(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create GitHub client: %w", err)
		}
		return &Client{
			github: github.NewClient(http.DefaultClient),
			ctx:    context.Background(),
			opts:   &ClientOptions{AuthMethod: AuthMethodPAT},
		}, nil
	}

	switch opts.AuthMethod {
	case AuthMethodPAT:
		_, err := gh.RESTClient(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create GitHub client: %w", err)
		}
		return &Client{
			github: github.NewClient(http.DefaultClient),
			ctx:    context.Background(),
			opts:   opts,
		}, nil

	case AuthMethodGitHubApp:
		client := &Client{
			ctx:  context.Background(),
			opts: opts,
		}

		// Get initial token
		if err := client.refreshToken(); err != nil {
			return nil, fmt.Errorf("failed to get initial token: %w", err)
		}

		return client, nil

	default:
		return nil, fmt.Errorf("unsupported authentication method")
	}
}

func (c *Client) refreshToken() error {
	if c.opts.AuthServer == "" {
		return fmt.Errorf("auth server URL is required")
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/token", c.opts.AuthServer), nil)
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	q := req.URL.Query()
	q.Add("app_id", fmt.Sprintf("%d", c.opts.AppID))
	q.Add("installation_id", fmt.Sprintf("%d", c.opts.InstallationID))
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get token from auth server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth server returned status %d", resp.StatusCode)
	}

	var authResp authResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
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
	return http.DefaultTransport.RoundTrip(req)
}

func (c *Client) ensureValidToken() error {
	if c.opts.AuthMethod != AuthMethodGitHubApp {
		return nil
	}

	// Refresh token if it's expired or will expire in the next minute
	if time.Now().Add(time.Minute).After(c.expiresAt) {
		if err := c.refreshToken(); err != nil {
			return fmt.Errorf("failed to refresh token: %w", err)
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

	secrets, _, err := c.github.Actions.ListOrgSecrets(c.ctx, org, &github.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list organization secrets: %w", err)
	}
	return secrets.Secrets, nil
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

	_, err = c.github.Actions.CreateOrUpdateOrgSecret(c.ctx, org, encryptedSecret)
	return err
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

	_, err = c.github.Actions.CreateOrUpdateRepoSecret(c.ctx, owner, repo, encryptedSecret)
	return err
}

func (c *Client) DeleteOrgSecret(org, secretName string) error {
	if err := c.ensureValidToken(); err != nil {
		return err
	}

	_, err := c.github.Actions.DeleteOrgSecret(c.ctx, org, secretName)
	if err != nil {
		return fmt.Errorf("failed to delete organization secret: %w", err)
	}
	return nil
}

func (c *Client) DeleteRepoSecret(owner, repo, secretName string) error {
	if err := c.ensureValidToken(); err != nil {
		return err
	}

	_, err := c.github.Actions.DeleteRepoSecret(c.ctx, owner, repo, secretName)
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
	req, err := c.github.NewRequest("PUT", url, encryptedSecret)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = c.github.Do(c.ctx, req, nil)
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
