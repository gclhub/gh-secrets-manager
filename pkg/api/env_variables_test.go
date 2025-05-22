package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestEnvVariablesCRUD(t *testing.T) {
	// Setup test server with handlers for each API endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/testorg/testrepo/environments/testenv/variables":
			if r.Method == "GET" {
				// Return list of variables
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"variables": []map[string]string{
						{"name": "EXISTING_VAR", "value": "existing-value"},
					},
				})
			} else if r.Method == "PUT" {
				// Create/update variable
				w.WriteHeader(http.StatusNoContent)
			}
		case "/repos/testorg/testrepo/environments/testenv/variables/TEST_VAR":
			if r.Method == "DELETE" {
				// Delete variable
				w.WriteHeader(http.StatusNoContent)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	
	// Create client pointing to our test server
	baseURL, _ := url.Parse(server.URL + "/")
	httpClient := newTestClient(baseURL)
	
	client := &Client{
		github: nil, // Don't need a real GitHub client for this test
		ctx:    nil,
		opts:   &ClientOptions{AuthMethod: AuthMethodPAT},
	}
	
	// Override the client's HTTP client with our test client
	// We're doing this in a way that avoids the need to import github package
	client.github = &clientWithHTTPClient{httpClient: httpClient}
	
	// Test listing variables
	variables, err := client.ListEnvironmentVariables("testorg", "testrepo", "testenv")
	if err != nil {
		t.Fatalf("ListEnvironmentVariables returned error: %v", err)
	}
	if len(variables) != 1 || variables[0].Name != "EXISTING_VAR" {
		t.Errorf("ListEnvironmentVariables returned incorrect variables: %+v", variables)
	}
	
	// Test creating/updating a variable
	variable := &Variable{Name: "TEST_VAR", Value: "test-value"}
	err = client.CreateOrUpdateEnvironmentVariable("testorg", "testrepo", "testenv", variable)
	if err != nil {
		t.Fatalf("CreateOrUpdateEnvironmentVariable returned error: %v", err)
	}
	
	// Test deleting a variable
	err = client.DeleteEnvironmentVariable("testorg", "testrepo", "testenv", "TEST_VAR")
	if err != nil {
		t.Fatalf("DeleteEnvironmentVariable returned error: %v", err)
	}
	
	// Test backward compatibility aliases
	variables, err = client.ListEnvVariables("testorg", "testrepo", "testenv")
	if err != nil {
		t.Fatalf("ListEnvVariables returned error: %v", err)
	}
	if len(variables) != 1 || variables[0].Name != "EXISTING_VAR" {
		t.Errorf("ListEnvVariables returned incorrect variables: %+v", variables)
	}
	
	err = client.CreateOrUpdateEnvVariable("testorg", "testrepo", "testenv", variable)
	if err != nil {
		t.Fatalf("CreateOrUpdateEnvVariable returned error: %v", err)
	}
	
	err = client.DeleteEnvVariable("testorg", "testrepo", "testenv", "TEST_VAR")
	if err != nil {
		t.Fatalf("DeleteEnvVariable returned error: %v", err)
	}
}

// Helper type to mock the GitHub client
type clientWithHTTPClient struct {
	httpClient *http.Client
}

// NewRequest implements the necessary method for the Client to use
func (c *clientWithHTTPClient) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	req, err := http.NewRequest(method, urlStr, nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}

// Do implements the necessary method for the Client to use
func (c *clientWithHTTPClient) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	
	if v != nil && resp.StatusCode == http.StatusOK {
		json.NewDecoder(resp.Body).Decode(v)
	}
	
	return resp, nil
}
