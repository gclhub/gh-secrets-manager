package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v45/github"
)

func setupCRUDTestServer() (*httptest.Server, *Client) {
	// Setup a test server that handles various API endpoints
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set default content type for all responses
		w.Header().Set("Content-Type", "application/json")
		
		// Handle different endpoints
		switch {
		// Organization secrets
		case r.URL.Path == "/orgs/testorg/actions/secrets":
			if r.Method == "GET" {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"secrets": []map[string]string{
						{"name": "ORG_SECRET1"},
						{"name": "ORG_SECRET2"},
					},
				})
			}
		case r.URL.Path == "/orgs/testorg/actions/secrets/TEST_SECRET":
			if r.Method == "PUT" || r.Method == "DELETE" {
				w.WriteHeader(http.StatusNoContent)
			}
		
		// Repository secrets
		case r.URL.Path == "/repos/testorg/testrepo/actions/secrets":
			if r.Method == "GET" {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"secrets": []map[string]string{
						{"name": "REPO_SECRET1"},
						{"name": "REPO_SECRET2"},
					},
				})
			}
		case r.URL.Path == "/repos/testorg/testrepo/actions/secrets/TEST_SECRET":
			if r.Method == "PUT" || r.Method == "DELETE" {
				w.WriteHeader(http.StatusNoContent)
			}
		
		// Public keys
		case r.URL.Path == "/orgs/testorg/actions/secrets/public-key":
			json.NewEncoder(w).Encode(map[string]string{
				"key_id": "test-key-id",
				"key":    "MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=", // base64 encoded 32 bytes
			})
		case r.URL.Path == "/repos/testorg/testrepo/actions/secrets/public-key":
			json.NewEncoder(w).Encode(map[string]string{
				"key_id": "test-key-id",
				"key":    "MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=", // base64 encoded 32 bytes
			})
		
		// Organization variables
		case r.URL.Path == "/orgs/testorg/actions/variables":
			if r.Method == "GET" {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"variables": []map[string]string{
						{"name": "ORG_VAR1", "value": "value1"},
						{"name": "ORG_VAR2", "value": "value2"},
					},
				})
			}
		case r.URL.Path == "/orgs/testorg/actions/variables/TEST_VAR":
			if r.Method == "POST" || r.Method == "PATCH" || r.Method == "DELETE" {
				w.WriteHeader(http.StatusNoContent)
			}
		
		// Repository variables
		case r.URL.Path == "/repos/testorg/testrepo/actions/variables":
			if r.Method == "GET" {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"variables": []map[string]string{
						{"name": "REPO_VAR1", "value": "value1"},
						{"name": "REPO_VAR2", "value": "value2"},
					},
				})
			}
		case r.URL.Path == "/repos/testorg/testrepo/actions/variables/TEST_VAR":
			if r.Method == "POST" || r.Method == "PATCH" || r.Method == "DELETE" {
				w.WriteHeader(http.StatusNoContent)
			}
		
		// Dependabot secrets
		case r.URL.Path == "/orgs/testorg/dependabot/secrets":
			if r.Method == "GET" {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"secrets": []map[string]string{
						{"name": "ORG_DEPENDABOT_SECRET1"},
						{"name": "ORG_DEPENDABOT_SECRET2"},
					},
				})
			}
		case r.URL.Path == "/orgs/testorg/dependabot/secrets/TEST_SECRET":
			if r.Method == "PUT" || r.Method == "DELETE" {
				w.WriteHeader(http.StatusNoContent)
			}
		case r.URL.Path == "/repos/testorg/testrepo/dependabot/secrets":
			if r.Method == "GET" {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"secrets": []map[string]string{
						{"name": "REPO_DEPENDABOT_SECRET1"},
						{"name": "REPO_DEPENDABOT_SECRET2"},
					},
				})
			}
		case r.URL.Path == "/repos/testorg/testrepo/dependabot/secrets/TEST_SECRET":
			if r.Method == "PUT" || r.Method == "DELETE" {
				w.WriteHeader(http.StatusNoContent)
			}
		case r.URL.Path == "/orgs/testorg/dependabot/secrets/public-key":
			json.NewEncoder(w).Encode(map[string]string{
				"key_id": "test-key-id",
				"key":    "MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=", // base64 encoded 32 bytes
			})
		case r.URL.Path == "/repos/testorg/testrepo/dependabot/secrets/public-key":
			json.NewEncoder(w).Encode(map[string]string{
				"key_id": "test-key-id",
				"key":    "MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=", // base64 encoded 32 bytes
			})
		
		// Environment secrets and variables
		case r.URL.Path == "/repos/testorg/testrepo/environments/testenv/secrets":
			if r.Method == "GET" {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"secrets": []map[string]string{
						{"name": "ENV_SECRET1"},
						{"name": "ENV_SECRET2"},
					},
				})
			}
		case r.URL.Path == "/repos/testorg/testrepo/environments/testenv/secrets/TEST_SECRET":
			if r.Method == "PUT" || r.Method == "DELETE" {
				w.WriteHeader(http.StatusNoContent)
			}
		case r.URL.Path == "/repos/testorg/testrepo/environments/testenv/secrets/public-key":
			json.NewEncoder(w).Encode(map[string]string{
				"key_id": "test-key-id",
				"key":    "MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=", // base64 encoded 32 bytes
			})
		case r.URL.Path == "/repos/testorg/testrepo/environments/testenv/variables":
			if r.Method == "GET" {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"variables": []map[string]string{
						{"name": "ENV_VAR1", "value": "value1"},
						{"name": "ENV_VAR2", "value": "value2"},
					},
				})
			}
		case r.URL.Path == "/repos/testorg/testrepo/environments/testenv/variables/TEST_VAR":
			if r.Method == "POST" || r.Method == "PATCH" || r.Method == "DELETE" {
				w.WriteHeader(http.StatusNoContent)
			}
		
		default:
			fmt.Printf("Unhandled path: %s %s\n", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	
	// Create client pointing to our test server
	baseURL, _ := url.Parse(server.URL + "/")
	httpClient := newTestClient(baseURL)
	
	client := &Client{
		github: github.NewClient(httpClient),
		ctx:    context.Background(),
		opts:   &ClientOptions{AuthMethod: AuthMethodPAT},
	}
	
	return server, client
}

func TestOrgSecretsCRUD(t *testing.T) {
	server, client := setupCRUDTestServer()
	defer server.Close()
	
	// Test listing secrets
	secrets, err := client.ListOrgSecrets("testorg")
	if err != nil {
		t.Fatalf("ListOrgSecrets returned error: %v", err)
	}
	
	if len(secrets) != 2 {
		t.Errorf("ListOrgSecrets returned %d secrets, want 2", len(secrets))
	}
	
	// Test creating/updating a secret
	secret := &github.EncryptedSecret{
		Name:           "TEST_SECRET",
		KeyID:          "test-key-id",
		EncryptedValue: "encrypted-value",
	}
	
	err = client.CreateOrUpdateOrgSecret("testorg", secret)
	if err != nil {
		t.Fatalf("CreateOrUpdateOrgSecret returned error: %v", err)
	}
	
	// Test deleting a secret
	err = client.DeleteOrgSecret("testorg", "TEST_SECRET")
	if err != nil {
		t.Fatalf("DeleteOrgSecret returned error: %v", err)
	}
}

func TestRepoSecretsCRUD(t *testing.T) {
	server, client := setupCRUDTestServer()
	defer server.Close()
	
	// Test listing secrets
	secrets, err := client.ListRepoSecrets("testorg", "testrepo")
	if err != nil {
		t.Fatalf("ListRepoSecrets returned error: %v", err)
	}
	
	if len(secrets) != 2 {
		t.Errorf("ListRepoSecrets returned %d secrets, want 2", len(secrets))
	}
	
	// Test creating/updating a secret
	secret := &github.EncryptedSecret{
		Name:           "TEST_SECRET",
		KeyID:          "test-key-id",
		EncryptedValue: "encrypted-value",
	}
	
	err = client.CreateOrUpdateRepoSecret("testorg", "testrepo", secret)
	if err != nil {
		t.Fatalf("CreateOrUpdateRepoSecret returned error: %v", err)
	}
	
	// Test deleting a secret
	err = client.DeleteRepoSecret("testorg", "testrepo", "TEST_SECRET")
	if err != nil {
		t.Fatalf("DeleteRepoSecret returned error: %v", err)
	}
}

func TestOrgVariablesCRUD(t *testing.T) {
	server, client := setupCRUDTestServer()
	defer server.Close()
	
	// Test listing variables
	variables, err := client.ListOrgVariables("testorg")
	if err != nil {
		t.Fatalf("ListOrgVariables returned error: %v", err)
	}
	
	if len(variables) != 2 {
		t.Errorf("ListOrgVariables returned %d variables, want 2", len(variables))
	}
	
	// Test creating/updating a variable
	variable := &Variable{
		Name:  "TEST_VAR",
		Value: "test-value",
	}
	
	err = client.CreateOrUpdateOrgVariable("testorg", variable)
	if err != nil {
		t.Fatalf("CreateOrUpdateOrgVariable returned error: %v", err)
	}
	
	// Test deleting a variable
	err = client.DeleteOrgVariable("testorg", "TEST_VAR")
	if err != nil {
		t.Fatalf("DeleteOrgVariable returned error: %v", err)
	}
}

func TestRepoVariablesCRUD(t *testing.T) {
	server, client := setupCRUDTestServer()
	defer server.Close()
	
	// Test listing variables
	variables, err := client.ListRepoVariables("testorg", "testrepo")
	if err != nil {
		t.Fatalf("ListRepoVariables returned error: %v", err)
	}
	
	if len(variables) != 2 {
		t.Errorf("ListRepoVariables returned %d variables, want 2", len(variables))
	}
	
	// Test creating/updating a variable
	variable := &Variable{
		Name:  "TEST_VAR",
		Value: "test-value",
	}
	
	err = client.CreateOrUpdateRepoVariable("testorg", "testrepo", variable)
	if err != nil {
		t.Fatalf("CreateOrUpdateRepoVariable returned error: %v", err)
	}
	
	// Test deleting a variable
	err = client.DeleteRepoVariable("testorg", "testrepo", "TEST_VAR")
	if err != nil {
		t.Fatalf("DeleteRepoVariable returned error: %v", err)
	}
}

func TestDependabotSecretsCRUD(t *testing.T) {
	server, client := setupCRUDTestServer()
	defer server.Close()
	
	// Test listing org dependabot secrets
	secrets, err := client.ListOrgDependabotSecrets("testorg")
	if err != nil {
		t.Fatalf("ListOrgDependabotSecrets returned error: %v", err)
	}
	
	if len(secrets) != 2 {
		t.Errorf("ListOrgDependabotSecrets returned %d secrets, want 2", len(secrets))
	}
	
	// Test creating/updating an org dependabot secret
	secret := &github.EncryptedSecret{
		Name:           "TEST_SECRET",
		KeyID:          "test-key-id",
		EncryptedValue: "encrypted-value",
	}
	
	err = client.CreateOrUpdateOrgDependabotSecret("testorg", secret)
	if err != nil {
		t.Fatalf("CreateOrUpdateOrgDependabotSecret returned error: %v", err)
	}
	
	// Test deleting an org dependabot secret
	err = client.DeleteOrgDependabotSecret("testorg", "TEST_SECRET")
	if err != nil {
		t.Fatalf("DeleteOrgDependabotSecret returned error: %v", err)
	}
	
	// Test listing repo dependabot secrets
	secrets, err = client.ListRepoDependabotSecrets("testorg", "testrepo")
	if err != nil {
		t.Fatalf("ListRepoDependabotSecrets returned error: %v", err)
	}
	
	if len(secrets) != 2 {
		t.Errorf("ListRepoDependabotSecrets returned %d secrets, want 2", len(secrets))
	}
	
	// Test creating/updating a repo dependabot secret
	err = client.CreateOrUpdateRepoDependabotSecret("testorg", "testrepo", secret)
	if err != nil {
		t.Fatalf("CreateOrUpdateRepoDependabotSecret returned error: %v", err)
	}
	
	// Test deleting a repo dependabot secret
	err = client.DeleteRepoDependabotSecret("testorg", "testrepo", "TEST_SECRET")
	if err != nil {
		t.Fatalf("DeleteRepoDependabotSecret returned error: %v", err)
	}
}

func TestEnvironmentSecretsCRUD(t *testing.T) {
	server, client := setupCRUDTestServer()
	defer server.Close()
	
	// Test getting environment public key
	pubKey, err := client.GetEnvironmentPublicKey("testorg", "testrepo", "testenv")
	if err != nil {
		t.Fatalf("GetEnvironmentPublicKey returned error: %v", err)
	}
	
	if pubKey.KeyID != "test-key-id" {
		t.Errorf("GetEnvironmentPublicKey returned KeyID %q, want %q", pubKey.KeyID, "test-key-id")
	}
	
	// Test listing environment secrets
	secrets, err := client.ListEnvironmentSecrets("testorg", "testrepo", "testenv")
	if err != nil {
		t.Fatalf("ListEnvironmentSecrets returned error: %v", err)
	}
	
	if len(secrets) != 2 {
		t.Errorf("ListEnvironmentSecrets returned %d secrets, want 2", len(secrets))
	}
	
	// Test creating/updating an environment secret
	secret := &github.EncryptedSecret{
		Name:           "TEST_SECRET",
		KeyID:          "test-key-id",
		EncryptedValue: "encrypted-value",
	}
	
	err = client.CreateOrUpdateEnvironmentSecret("testorg", "testrepo", "testenv", secret)
	if err != nil {
		t.Fatalf("CreateOrUpdateEnvironmentSecret returned error: %v", err)
	}
	
	// Test deleting an environment secret
	err = client.DeleteEnvironmentSecret("testorg", "testrepo", "testenv", "TEST_SECRET")
	if err != nil {
		t.Fatalf("DeleteEnvironmentSecret returned error: %v", err)
	}
	
	// Test alias methods
	secrets, err = client.ListEnvSecrets("testorg", "testrepo", "testenv")
	if err != nil {
		t.Fatalf("ListEnvSecrets returned error: %v", err)
	}
	
	if len(secrets) != 2 {
		t.Errorf("ListEnvSecrets returned %d secrets, want 2", len(secrets))
	}
	
	err = client.CreateOrUpdateEnvSecret("testorg", "testrepo", "testenv", secret)
	if err != nil {
		t.Fatalf("CreateOrUpdateEnvSecret returned error: %v", err)
	}
	
	err = client.DeleteEnvSecret("testorg", "testrepo", "testenv", "TEST_SECRET")
	if err != nil {
		t.Fatalf("DeleteEnvSecret returned error: %v", err)
	}
}

func TestGetPublicKeys(t *testing.T) {
	server, client := setupCRUDTestServer()
	defer server.Close()
	
	// Test getting org public key
	orgPubKey, err := client.GetOrgPublicKey("testorg")
	if err != nil {
		t.Fatalf("GetOrgPublicKey returned error: %v", err)
	}
	
	if orgPubKey.KeyID != "test-key-id" {
		t.Errorf("GetOrgPublicKey returned KeyID %q, want %q", orgPubKey.KeyID, "test-key-id")
	}
	
	// Test getting repo public key
	repoPubKey, err := client.GetRepoPublicKey("testorg", "testrepo")
	if err != nil {
		t.Fatalf("GetRepoPublicKey returned error: %v", err)
	}
	
	if repoPubKey.KeyID != "test-key-id" {
		t.Errorf("GetRepoPublicKey returned KeyID %q, want %q", repoPubKey.KeyID, "test-key-id")
	}
}
