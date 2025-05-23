package api

import (
	"testing"
)

func TestGetOrgPublicKey_ValidResponse(t *testing.T) {
	// Setup test server
	response := struct {
		Key   string `json:"key"`
		KeyID string `json:"key_id"`
	}{
		Key:   "dGVzdC1rZXk=", // base64 encoded "test-key"
		KeyID: "test-key-id",
	}
	server, client := setupTestServer(t, "/orgs/testorg/actions/secrets/public-key", response)
	defer server.Close()

	// Call the method
	encryption, err := client.GetOrgPublicKey("testorg")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify the response
	if encryption.KeyID != response.KeyID {
		t.Errorf("Expected key ID %q, got %q", response.KeyID, encryption.KeyID)
	}
	if len(encryption.PublicKey) != 8 { // "test-key" is 8 bytes when decoded
		t.Errorf("Expected public key length 8, got %d", len(encryption.PublicKey))
	}
}

func TestGetRepoPublicKey_ValidResponse(t *testing.T) {
	// Setup test server
	response := struct {
		Key   string `json:"key"`
		KeyID string `json:"key_id"`
	}{
		Key:   "dGVzdC1rZXk=", // base64 encoded "test-key"
		KeyID: "test-key-id",
	}
	server, client := setupTestServer(t, "/repos/testorg/testrepo/actions/secrets/public-key", response)
	defer server.Close()

	// Call the method
	encryption, err := client.GetRepoPublicKey("testorg", "testrepo")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify the response
	if encryption.KeyID != response.KeyID {
		t.Errorf("Expected key ID %q, got %q", response.KeyID, encryption.KeyID)
	}
	if len(encryption.PublicKey) != 8 { // "test-key" is 8 bytes when decoded
		t.Errorf("Expected public key length 8, got %d", len(encryption.PublicKey))
	}
}

func TestGetEnvironmentPublicKey_ValidResponse(t *testing.T) {
	// Setup test server
	response := struct {
		Key   string `json:"key"`
		KeyID string `json:"key_id"`
	}{
		Key:   "dGVzdC1rZXk=", // base64 encoded "test-key"
		KeyID: "test-key-id",
	}
	server, client := setupTestServer(t, "/repos/testorg/testrepo/environments/testenv/secrets/public-key", response)
	defer server.Close()

	// Call the method
	encryption, err := client.GetEnvironmentPublicKey("testorg", "testrepo", "testenv")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify the response
	if encryption.KeyID != response.KeyID {
		t.Errorf("Expected key ID %q, got %q", response.KeyID, encryption.KeyID)
	}
	if len(encryption.PublicKey) != 8 { // "test-key" is 8 bytes when decoded
		t.Errorf("Expected public key length 8, got %d", len(encryption.PublicKey))
	}
}