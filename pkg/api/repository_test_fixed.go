package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v45/github"
)

func TestListRepositoriesByProperty(t *testing.T) {
	// Setup test server with handlers for repository API endpoints
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		if r.URL.Path == "/orgs/testorg/properties/values" && r.URL.Query().Get("property_name") == "team" && r.URL.Query().Get("value") == "backend" {
			// Return list of repositories with the property
			json.NewEncoder(w).Encode(map[string]interface{}{
				"repository_names": []string{
					"repo1",
					"repo2",
				},
				"total_count": 2,
				"has_next_page": false,
			})
		} else if r.URL.Path == "/repos/testorg/repo1" {
			// Return repository details
			json.NewEncoder(w).Encode(map[string]interface{}{
				"name": "repo1",
				"full_name": "testorg/repo1",
				"owner": map[string]interface{}{
					"login": "testorg",
				},
			})
		} else if r.URL.Path == "/repos/testorg/repo2" {
			// Return repository details
			json.NewEncoder(w).Encode(map[string]interface{}{
				"name": "repo2",
				"full_name": "testorg/repo2",
				"owner": map[string]interface{}{
					"login": "testorg",
				},
			})
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	
	// Create client pointing to our test server
	baseURL, _ := url.Parse(server.URL + "/")
	httpClient := newTestClient(baseURL)
	
	client := &Client{
		github: github.NewClient(httpClient),
		ctx:    context.Background(),
		opts:   &ClientOptions{AuthMethod: AuthMethodPAT},
	}
	
	// Test listing repositories by property
	repos, err := client.ListRepositoriesByProperty("testorg", "team", "backend")
	if err != nil {
		t.Fatalf("ListRepositoriesByProperty returned error: %v", err)
	}
	
	// Check the results
	if len(repos) != 2 {
		t.Errorf("ListRepositoriesByProperty returned %d repos, want 2", len(repos))
		return
	}
	
	expected := map[string]bool{
		"repo1": false,
		"repo2": false,
	}
	
	for _, repo := range repos {
		name := *repo.Name
		if _, ok := expected[name]; ok {
			expected[name] = true
		} else {
			t.Errorf("Unexpected repository: %s", name)
		}
	}
	
	for name, found := range expected {
		if !found {
			t.Errorf("Expected repository %s was not returned", name)
		}
	}
}

func TestListRepositoriesByProperty_Error(t *testing.T) {
	client := &Client{
		ctx: context.Background(),
	}
	
	// Test with empty org
	_, err := client.ListRepositoriesByProperty("", "team", "backend")
	if err == nil {
		t.Error("ListRepositoriesByProperty with empty org should return error")
	}
	
	// Test with empty property name
	_, err = client.ListRepositoriesByProperty("testorg", "", "backend")
	if err == nil {
		t.Error("ListRepositoriesByProperty with empty property name should return error")
	}
	
	// Test with empty property value
	_, err = client.ListRepositoriesByProperty("testorg", "team", "")
	if err == nil {
		t.Error("ListRepositoriesByProperty with empty property value should return error")
	}
	
	// Test with server error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()
	
	baseURL, _ := url.Parse(server.URL + "/")
	httpClient := newTestClient(baseURL)
	
	client = &Client{
		github: github.NewClient(httpClient),
		ctx:    context.Background(),
		opts:   &ClientOptions{AuthMethod: AuthMethodPAT},
	}
	
	_, err = client.ListRepositoriesByProperty("testorg", "team", "backend")
	if err == nil {
		t.Error("ListRepositoriesByProperty with server error should return error")
	}
}
