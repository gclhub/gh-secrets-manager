package api

import (
	"encoding/base64"
	"testing"
)

func TestSecretEncryption_EncryptSecret(t *testing.T) {
	// Create a valid 32-byte key
	validKey := make([]byte, 32)
	for i := 0; i < 32; i++ {
		validKey[i] = byte(i)
	}
	
	encryption := &SecretEncryption{
		KeyID:     "test-key-id",
		PublicKey: validKey,
	}
	
	// Test encrypting a secret
	encrypted, err := encryption.EncryptSecret("test-secret-value")
	if err != nil {
		t.Fatalf("EncryptSecret returned error: %v", err)
	}
	
	// Encrypted value should be non-empty and base64-encoded
	if encrypted == "" {
		t.Errorf("EncryptSecret returned empty string")
	}
	
	// Should be valid base64
	_, err = base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		t.Errorf("EncryptSecret returned invalid base64: %v", err)
	}
}

func TestSecretEncryption_EncryptSecret_InvalidKey(t *testing.T) {
	// Create an invalid key (not 32 bytes)
	invalidKey := make([]byte, 16)
	
	encryption := &SecretEncryption{
		KeyID:     "test-key-id",
		PublicKey: invalidKey,
	}
	
	// This should return an error
	_, err := encryption.EncryptSecret("test-secret-value")
	if err == nil {
		t.Error("EncryptSecret with invalid key should return error")
	}
}

func TestSecretEncryption_CreateEncryptedSecret(t *testing.T) {
	// Create a valid 32-byte key
	validKey := make([]byte, 32)
	for i := 0; i < 32; i++ {
		validKey[i] = byte(i)
	}
	
	encryption := &SecretEncryption{
		KeyID:     "test-key-id",
		PublicKey: validKey,
	}
	
	// Test creating an encrypted secret
	secret, err := encryption.CreateEncryptedSecret("TEST_SECRET", "test-secret-value")
	if err != nil {
		t.Fatalf("CreateEncryptedSecret returned error: %v", err)
	}
	
	// Check the secret properties
	if secret.Name != "TEST_SECRET" {
		t.Errorf("Secret name = %q, want %q", secret.Name, "TEST_SECRET")
	}
	
	if secret.KeyID != "test-key-id" {
		t.Errorf("Secret KeyID = %q, want %q", secret.KeyID, "test-key-id")
	}
	
	if secret.EncryptedValue == "" {
		t.Error("Secret EncryptedValue is empty")
	}
	
	// Should be valid base64
	_, err = base64.StdEncoding.DecodeString(secret.EncryptedValue)
	if err != nil {
		t.Errorf("Secret EncryptedValue is invalid base64: %v", err)
	}
}
