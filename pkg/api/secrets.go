package api

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/google/go-github/v45/github"
	"golang.org/x/crypto/nacl/box"
)

// SecretEncryption handles the encryption of secrets using GitHub's public key
type SecretEncryption struct {
	KeyID     string
	PublicKey []byte
}

// GetOrgPublicKey fetches the public key for encrypting organization secrets
func (c *Client) GetOrgPublicKey(org string) (*SecretEncryption, error) {
	if err := c.ensureValidToken(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("orgs/%s/actions/secrets/public-key", org)
	fmt.Printf("[DEBUG] GetOrgPublicKey creating request for URL: %s\n", url)
	req, err := c.github.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var key struct {
		KeyID string `json:"key_id"`
		Key   string `json:"key"`
	}
	fmt.Printf("[DEBUG] GetOrgPublicKey sending request: %s %s\n", req.Method, req.URL.String())
	_, err = c.github.Do(c.ctx, req, &key)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization public key: %w", err)
	}
	fmt.Printf("[DEBUG] GetOrgPublicKey got response: KeyID=%s, Key=%s\n", key.KeyID, key.Key)

	publicKey, err := base64.StdEncoding.DecodeString(key.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}
	fmt.Printf("[DEBUG] GetOrgPublicKey decoded key length: %d\n", len(publicKey))

	return &SecretEncryption{
		KeyID:     key.KeyID,
		PublicKey: publicKey,
	}, nil
}

// GetRepoPublicKey fetches the public key for encrypting repository secrets
func (c *Client) GetRepoPublicKey(owner, repo string) (*SecretEncryption, error) {
	if err := c.ensureValidToken(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("repos/%s/%s/actions/secrets/public-key", owner, repo)
	fmt.Printf("[DEBUG] GetRepoPublicKey creating request for URL: %s\n", url)
	req, err := c.github.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var key struct {
		KeyID string `json:"key_id"`
		Key   string `json:"key"`
	}
	fmt.Printf("[DEBUG] GetRepoPublicKey sending request: %s %s\n", req.Method, req.URL.String())
	_, err = c.github.Do(c.ctx, req, &key)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository public key: %w", err)
	}
	fmt.Printf("[DEBUG] GetRepoPublicKey got response: KeyID=%s, Key=%s\n", key.KeyID, key.Key)

	publicKey, err := base64.StdEncoding.DecodeString(key.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}
	fmt.Printf("[DEBUG] GetRepoPublicKey decoded key length: %d\n", len(publicKey))

	return &SecretEncryption{
		KeyID:     key.KeyID,
		PublicKey: publicKey,
	}, nil
}

// GetOrgDependabotPublicKey fetches the public key for encrypting organization Dependabot secrets
func (c *Client) GetOrgDependabotPublicKey(org string) (*SecretEncryption, error) {
	if err := c.ensureValidToken(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("orgs/%s/dependabot/secrets/public-key", org)
	fmt.Printf("[DEBUG] GetOrgDependabotPublicKey creating request for URL: %s\n", url)
	req, err := c.github.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var key struct {
		KeyID string `json:"key_id"`
		Key   string `json:"key"`
	}
	fmt.Printf("[DEBUG] GetOrgDependabotPublicKey sending request: %s %s\n", req.Method, req.URL.String())
	_, err = c.github.Do(c.ctx, req, &key)
	if err != nil {
		return nil, fmt.Errorf("failed to get org dependabot public key: %w", err)
	}
	fmt.Printf("[DEBUG] GetOrgDependabotPublicKey got response: KeyID=%s, Key=%s\n", key.KeyID, key.Key)

	publicKey, err := base64.StdEncoding.DecodeString(key.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}
	fmt.Printf("[DEBUG] GetOrgDependabotPublicKey decoded key length: %d\n", len(publicKey))

	return &SecretEncryption{
		KeyID:     key.KeyID,
		PublicKey: publicKey,
	}, nil
}

// GetRepoDependabotPublicKey fetches the public key for encrypting repository Dependabot secrets
func (c *Client) GetRepoDependabotPublicKey(owner, repo string) (*SecretEncryption, error) {
	if err := c.ensureValidToken(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("repos/%s/%s/dependabot/secrets/public-key", owner, repo)
	fmt.Printf("[DEBUG] GetRepoDependabotPublicKey creating request for URL: %s\n", url)
	req, err := c.github.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var key struct {
		KeyID string `json:"key_id"`
		Key   string `json:"key"`
	}
	fmt.Printf("[DEBUG] GetRepoDependabotPublicKey sending request: %s %s\n", req.Method, req.URL.String())
	_, err = c.github.Do(c.ctx, req, &key)
	if err != nil {
		return nil, fmt.Errorf("failed to get repo dependabot public key: %w", err)
	}
	fmt.Printf("[DEBUG] GetRepoDependabotPublicKey got response: KeyID=%s, Key=%s\n", key.KeyID, key.Key)

	publicKey, err := base64.StdEncoding.DecodeString(key.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}
	fmt.Printf("[DEBUG] GetRepoDependabotPublicKey decoded key length: %d\n", len(publicKey))

	return &SecretEncryption{
		KeyID:     key.KeyID,
		PublicKey: publicKey,
	}, nil
}

// GetEnvironmentPublicKey fetches the public key for encrypting environment secrets
func (c *Client) GetEnvironmentPublicKey(owner, repo, environment string) (*SecretEncryption, error) {
	if err := c.ensureValidToken(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("repos/%s/%s/environments/%s/secrets/public-key", owner, repo, environment)
	req, err := c.github.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var key struct {
		KeyID string `json:"key_id"`
		Key   string `json:"key"`
	}
	_, err = c.github.Do(c.ctx, req, &key)
	if err != nil {
		return nil, fmt.Errorf("failed to get environment public key: %w", err)
	}

	publicKey, err := base64.StdEncoding.DecodeString(key.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	return &SecretEncryption{
		KeyID:     key.KeyID,
		PublicKey: publicKey,
	}, nil
}

// EncryptSecret encrypts a secret value using libsodium's sealed box
func (s *SecretEncryption) EncryptSecret(secret string) (string, error) {
	if len(s.PublicKey) != 32 {
		return "", fmt.Errorf("invalid public key length: expected 32 bytes, got %d", len(s.PublicKey))
	}

	var publicKey [32]byte
	copy(publicKey[:], s.PublicKey)

	encrypted, err := box.SealAnonymous(nil, []byte(secret), &publicKey, rand.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt secret: %w", err)
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// CreateEncryptedSecret creates an EncryptedSecret with the given name and encrypted value
func (s *SecretEncryption) CreateEncryptedSecret(name, value string) (*github.EncryptedSecret, error) {
	encryptedValue, err := s.EncryptSecret(value)
	if err != nil {
		return nil, err
	}

	return &github.EncryptedSecret{
		Name:           name,
		KeyID:          s.KeyID,
		EncryptedValue: encryptedValue,
	}, nil
}

// CreateOrUpdateEnvironmentSecret creates or updates an environment-level secret
func (c *Client) CreateOrUpdateEnvironmentSecret(owner, repo, environment string, secret *github.EncryptedSecret) error {
	if err := c.ensureValidToken(); err != nil {
		return err
	}

	encryption, err := c.GetEnvironmentPublicKey(owner, repo, environment)
	if err != nil {
		return err
	}

	encryptedSecret, err := encryption.CreateEncryptedSecret(secret.Name, secret.EncryptedValue)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("repos/%s/%s/environments/%s/secrets/%s", owner, repo, environment, secret.Name)
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
		return fmt.Errorf("failed to create/update environment secret: %w", err)
	}

	return nil
}

// DeleteEnvironmentSecret deletes an environment-level secret
func (c *Client) DeleteEnvironmentSecret(owner, repo, environment, name string) error {
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

// ListEnvironmentSecrets lists all secrets available in an environment
func (c *Client) ListEnvironmentSecrets(owner, repo, environment string) ([]*github.Secret, error) {
	if err := c.ensureValidToken(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("repos/%s/%s/environments/%s/secrets", owner, repo, environment)
	req, err := c.github.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var response struct {
		Secrets []*github.Secret `json:"secrets"`
	}
	_, err = c.github.Do(c.ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list environment secrets: %w", err)
	}

	return response.Secrets, nil
}

// GetEnvironmentSecret gets a single environment-level secret
func (c *Client) GetEnvironmentSecret(owner, repo, environment, name string) (*github.Secret, error) {
	if err := c.ensureValidToken(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("repos/%s/%s/environments/%s/secrets/%s", owner, repo, environment, name)
	req, err := c.github.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	secret := &github.Secret{}
	_, err = c.github.Do(c.ctx, req, secret)
	if err != nil {
		return nil, fmt.Errorf("failed to get environment secret: %w", err)
	}

	return secret, nil
}
