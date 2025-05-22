package api

import (
	"testing"
)

func TestEnvironmentAliases(t *testing.T) {
	server, client := setupCRUDTestServer()
	defer server.Close()
	
	// Test ListEnvVariables alias
	variables, err := client.ListEnvVariables("testorg", "testrepo", "testenv")
	if err != nil {
		t.Fatalf("ListEnvVariables returned error: %v", err)
	}
	
	if len(variables) != 2 {
		t.Errorf("ListEnvVariables returned %d variables, want 2", len(variables))
	}
	
	// Test CreateOrUpdateEnvVariable alias
	variable := &Variable{
		Name:  "TEST_VAR",
		Value: "test-value",
	}
	
	err = client.CreateOrUpdateEnvVariable("testorg", "testrepo", "testenv", variable)
	if err != nil {
		t.Fatalf("CreateOrUpdateEnvVariable returned error: %v", err)
	}
	
	// Test DeleteEnvVariable alias
	err = client.DeleteEnvVariable("testorg", "testrepo", "testenv", "TEST_VAR")
	if err != nil {
		t.Fatalf("DeleteEnvVariable returned error: %v", err)
	}
}

func TestGetEnvironmentSecret(t *testing.T) {
	// Setup test server with handler for GetEnvironmentSecret
	server, client := setupCRUDTestServer()
	defer server.Close()
	
	// Since this path is not handled in our test server, we should get "not found"
	_, err := client.GetEnvironmentSecret("testorg", "testrepo", "testenv", "NOT_FOUND_SECRET")
	if err == nil {
		t.Error("GetEnvironmentSecret with non-existent secret should return error")
	}
}
