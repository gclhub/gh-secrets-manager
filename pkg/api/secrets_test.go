package api

import (
	"encoding/base64"
	"testing"

	"github.com/google/go-github/v45/github"
)

func TestSecretEncryption(t *testing.T) {
	// Test GetOrgPublicKey
	expectedKey := &github.PublicKey{
		KeyID: github.String("test-key-id"),
		Key:   github.String("mDGgF4WqsUwb2bW2kgE4AltE382qHtGr2wIrC+JTdR4="), // Sample base64 encoded key
	}

	server, client := setupTestServer(t, "/orgs/testorg/actions/secrets/public-key", expectedKey)
	defer server.Close()

	encryption, err := client.GetOrgPublicKey("testorg")
	if err != nil {
		t.Fatalf("GetOrgPublicKey returned error: %v", err)
	}

	if encryption.KeyID != *expectedKey.KeyID {
		t.Errorf("KeyID = %v, want %v", encryption.KeyID, *expectedKey.KeyID)
	}

	// Test secret encryption
	testSecret := "test-secret-value"
	encrypted, err := encryption.EncryptSecret(testSecret)
	if err != nil {
		t.Fatalf("EncryptSecret returned error: %v", err)
	}

	// Verify the encrypted value is base64 encoded
	if _, err := base64.StdEncoding.DecodeString(encrypted); err != nil {
		t.Errorf("Encrypted value is not valid base64: %v", err)
	}

	// Test creating encrypted secret
	secret, err := encryption.CreateEncryptedSecret("TEST_SECRET", testSecret)
	if err != nil {
		t.Fatalf("CreateEncryptedSecret returned error: %v", err)
	}

	if secret.Name != "TEST_SECRET" {
		t.Errorf("Secret name = %v, want TEST_SECRET", secret.Name)
	}
	if secret.KeyID != encryption.KeyID {
		t.Errorf("Secret KeyID = %v, want %v", secret.KeyID, encryption.KeyID)
	}
	if _, err := base64.StdEncoding.DecodeString(secret.EncryptedValue); err != nil {
		t.Errorf("Secret EncryptedValue is not valid base64: %v", err)
	}
}