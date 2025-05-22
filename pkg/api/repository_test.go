package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestListRepositoriesByProperty(t *testing.T) {
	// Setup test server with handlers for repository API endpoints
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/orgs/testorg/properties/values" && r.URL.Query().Get("property_name") == "team" && r.URL.Query().Get("value") == "backend" {
			// Return list of repositories with the property
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"repositories": []map[string]string{
					{"name": "repo1"},
					{"name": "repo2"},
				},
				"has_next_page": false,
			})
		} else if r.URL.Path == "/repos/testorg/repo1" {
			// Return repository details
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"name": "repo1",
				"full_name": "testorg/repo1",
				"owner": map[string]string{
					"login": "testorg",
				},
			})
		} else if r.URL.Path == "/repos/testorg/repo2" {
			// Return repository details
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"name": "repo2",
				"full_name": "testorg/repo2",
				"owner": map[string]string{
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
	
	// Create a custom mockRepositoriesService that uses our HTTP client
	repoService := &mockRepositoriesService{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
	
	client := &Client{
		github: &mockGithubClient{
			repoService: repoService,
		},
		ctx:  nil,
		opts: &ClientOptions{AuthMethod: AuthMethodPAT},
	}
	
	// Test listing repositories by property
	repos, err := client.ListRepositoriesByProperty("testorg", "team", "backend")
	if err != nil {
		t.Fatalf("ListRepositoriesByProperty returned error: %v", err)
	}
	
	// Check the results
	if len(repos) != 2 {
		t.Errorf("ListRepositoriesByProperty returned %d repos, want 2", len(repos))
	}
	
	repoNames := map[string]bool{}
	for _, repo := range repos {
		repoNames[*repo.Name] = true
	}
	
	if !repoNames["repo1"] || !repoNames["repo2"] {
		t.Errorf("ListRepositoriesByProperty returned incorrect repos: %v", repoNames)
	}
}

func TestListRepositoriesByProperty_InvalidParams(t *testing.T) {
	client := &Client{}
	
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
}

// Mock types to support the tests
type mockGithubClient struct {
	repoService *mockRepositoriesService
}

func (m *mockGithubClient) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	return http.NewRequest(method, urlStr, nil)
}

func (m *mockGithubClient) Do(req *http.Request, v interface{}) (*http.Response, error) {
	// Use the repoService's client for HTTP calls
	resp, err := m.repoService.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	
	if v != nil && resp.StatusCode == http.StatusOK {
		json.NewDecoder(resp.Body).Decode(v)
	}
	
	return resp, nil
}

type mockRepositoriesService struct {
	httpClient *http.Client
	baseURL    *url.URL
}

func (m *mockRepositoriesService) Get(ctx interface{}, owner, repo string) (interface{}, interface{}, error) {
	req, _ := http.NewRequest("GET", m.baseURL.String()+"repos/"+owner+"/"+repo, nil)
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	
	var repoData map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&repoData)
	
	// Convert to Repository
	stringVal := func(m map[string]interface{}, key string) string {
		if v, ok := m[key].(string); ok {
			return v
		}
		return ""
	}
	
	name := stringVal(repoData, "name")
	fullName := stringVal(repoData, "full_name")
	
	return struct {
		Name     *string
		FullName *string
		Owner    struct {
			Login *string
		}
	}{
		Name:     &name,
		FullName: &fullName,
		Owner: struct {
			Login *string
		}{
			Login: func() *string {
				if owner, ok := repoData["owner"].(map[string]interface{}); ok {
					login := stringVal(owner, "login")
					return &login
				}
				return nil
			}(),
		},
	}, nil, nil
}
