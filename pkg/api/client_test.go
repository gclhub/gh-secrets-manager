package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/google/go-github/v45/github"
)

func TestMain(m *testing.M) {
	fmt.Println("Starting API tests...")
	code := m.Run()
	fmt.Println("Finished API tests.")
	os.Exit(code)
}

func setupTestServer(t *testing.T, path string, response interface{}) (*httptest.Server, *Client) {
	fmt.Printf("Setting up test server for path: %s\n", path)
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle custom properties API path format
		if path == "/orgs/testorg/properties/values" && strings.HasPrefix(r.URL.Path, path) {
			// Check query parameters
			q := r.URL.Query()
			if q.Get("property_name") == "" || q.Get("value") == "" {
				t.Errorf("Missing property_name or value query parameter")
				http.Error(w, "Missing required query parameters", http.StatusBadRequest)
				return
			}
		} else if r.URL.Path != path {
			t.Errorf("Expected path %q, got %q", path, r.URL.Path)
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))

	// Create a client that directly uses the test server
	httpClient := &http.Client{}
	client := &Client{
		github: github.NewClient(httpClient),
		ctx:    context.Background(),
		opts:   &ClientOptions{AuthMethod: AuthMethodPAT}, // Use PAT auth for testing
	}
	client.github.BaseURL, _ = url.Parse(server.URL + "/")

	return server, client
}

func TestListOrgSecrets(t *testing.T) {
	response := &github.Secrets{
		Secrets: []*github.Secret{
			{Name: "SECRET1"},
			{Name: "SECRET2"},
		},
	}

	server, client := setupTestServer(t, "/orgs/testorg/actions/secrets", response)
	defer server.Close()

	result, err := client.ListOrgSecrets("testorg")
	if err != nil {
		t.Fatalf("ListOrgSecrets returned error: %v", err)
	}

	if len(result) != len(response.Secrets) {
		t.Errorf("ListOrgSecrets returned %d secrets, want %d", len(result), len(response.Secrets))
	}

	for i, secret := range result {
		if secret.Name != response.Secrets[i].Name {
			t.Errorf("Secret %d: got name %q, want %q", i, secret.Name, response.Secrets[i].Name)
		}
	}
}

func TestListRepositoriesByProperty(t *testing.T) {
	type repoStruct struct {
		Name string `json:"name"`
	}

	type repoResponse struct {
		Repositories []repoStruct `json:"repositories"`
		HasNextPage  bool         `json:"has_next_page"`
	}

	t.Run("basic", func(t *testing.T) {
		propertyResponse := &repoResponse{
			Repositories: []repoStruct{
				{Name: "repo1"},
				{Name: "repo3"},
			},
			HasNextPage: false,
		}

		server1, client := setupTestServer(t, "/orgs/testorg/properties/values", propertyResponse)
		defer server1.Close()

		repos := map[string]*github.Repository{
			"repo1": {Name: github.String("repo1"), FullName: github.String("testorg/repo1")},
			"repo3": {Name: github.String("repo3"), FullName: github.String("testorg/repo3")},
		}

		server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			parts := strings.Split(r.URL.Path, "/")
			if len(parts) < 4 {
				t.Errorf("Invalid path: %s", r.URL.Path)
				http.NotFound(w, r)
				return
			}
			repoName := parts[len(parts)-1]
			if repo, ok := repos[repoName]; ok {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(repo)
				return
			}
			t.Errorf("Unexpected repository request: %s", repoName)
			http.NotFound(w, r)
		}))
		defer server2.Close()

		client.github.BaseURL, _ = url.Parse(server2.URL + "/")

		matchingRepos, err := client.ListRepositoriesByProperty("testorg", "team", "backend")
		if err != nil {
			t.Fatalf("ListRepositoriesByProperty returned error: %v", err)
		}

		if len(matchingRepos) != 2 {
			t.Errorf("ListRepositoriesByProperty returned %d repos, want 2", len(matchingRepos))
		}

		expectedMap := map[string]bool{"repo1": true, "repo3": true}
		for _, repo := range matchingRepos {
			if !expectedMap[repo.GetName()] {
				t.Errorf("Got unexpected repository %s", repo.GetName())
			}
		}
	})

	t.Run("with_pagination", func(t *testing.T) {
		var pageNum int
		responses := []*repoResponse{
			{
				Repositories: []repoStruct{{Name: "repo1"}},
				HasNextPage:  true,
			},
			{
				Repositories: []repoStruct{{Name: "repo3"}},
				HasNextPage:  false,
			},
		}

		server1, client := setupTestServer(t, "/orgs/testorg/properties/values", nil)
		server1.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.Path, "/orgs/testorg/properties/values") {
				t.Errorf("Unexpected path: %s", r.URL.Path)
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(responses[pageNum])
			if pageNum < len(responses)-1 {
				pageNum++
			}
		})
		defer server1.Close()

		repos := map[string]*github.Repository{
			"repo1": {Name: github.String("repo1"), FullName: github.String("testorg/repo1")},
			"repo3": {Name: github.String("repo3"), FullName: github.String("testorg/repo3")},
		}

		server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			parts := strings.Split(r.URL.Path, "/")
			if len(parts) < 4 {
				t.Errorf("Invalid path: %s", r.URL.Path)
				http.NotFound(w, r)
				return
			}
			repoName := parts[len(parts)-1]
			if repo, ok := repos[repoName]; ok {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(repo)
				return
			}
			t.Errorf("Unexpected repository request: %s", repoName)
			http.NotFound(w, r)
		}))
		defer server2.Close()

		client.github.BaseURL, _ = url.Parse(server2.URL + "/")

		matchingRepos, err := client.ListRepositoriesByProperty("testorg", "team", "backend")
		if err != nil {
			t.Fatalf("ListRepositoriesByProperty returned error: %v", err)
		}

		if len(matchingRepos) != 2 {
			t.Errorf("ListRepositoriesByProperty returned %d repos, want 2", len(matchingRepos))
		}

		expectedMap := map[string]bool{"repo1": true, "repo3": true}
		for _, repo := range matchingRepos {
			if !expectedMap[repo.GetName()] {
				t.Errorf("Got unexpected repository %s", repo.GetName())
			}
		}
	})
}

func TestListRepositoriesByPropertyErrors(t *testing.T) {
	t.Run("property_api_error", func(t *testing.T) {
		server, client := setupTestServer(t, "/orgs/testorg/properties/values", nil)
		server.Close() // Close immediately to simulate connection error

		_, err := client.ListRepositoriesByProperty("testorg", "team", "backend")
		if err == nil {
			t.Error("Expected error but got nil")
		}
	})

	t.Run("missing_org", func(t *testing.T) {
		server, client := setupTestServer(t, "/orgs/testorg/properties/values", nil)
		defer server.Close()

		_, err := client.ListRepositoriesByProperty("", "team", "backend")
		if err == nil {
			t.Error("Expected error but got nil")
		} else if !strings.Contains(err.Error(), "organization name") {
			t.Errorf("Expected error about missing organization name, got: %v", err)
		}
	})

	t.Run("missing_property_name", func(t *testing.T) {
		server, client := setupTestServer(t, "/orgs/testorg/properties/values", nil)
		defer server.Close()

		_, err := client.ListRepositoriesByProperty("testorg", "", "backend")
		if err == nil {
			t.Error("Expected error but got nil")
		} else if !strings.Contains(err.Error(), "property_name") {
			t.Errorf("Expected error about missing property_name, got: %v", err)
		}
	})

	t.Run("missing_property_value", func(t *testing.T) {
		server, client := setupTestServer(t, "/orgs/testorg/properties/values", nil)
		defer server.Close()

		_, err := client.ListRepositoriesByProperty("testorg", "team", "")
		if err == nil {
			t.Error("Expected error but got nil")
		} else if !strings.Contains(err.Error(), "value") {
			t.Errorf("Expected error about missing value, got: %v", err)
		}
	})

	t.Run("repository_api_error", func(t *testing.T) {
		type repoStruct struct {
			Name string `json:"name"`
		}

		propertyResponse := struct {
			Repositories []repoStruct `json:"repositories"`
			HasNextPage  bool         `json:"has_next_page"`
		}{
			Repositories: []repoStruct{{Name: "repo1"}},
			HasNextPage:  false,
		}

		server, client := setupTestServer(t, "/orgs/testorg/properties/values", propertyResponse)
		defer server.Close()

		_, err := client.ListRepositoriesByProperty("testorg", "team", "backend")
		if err == nil {
			t.Error("Expected error but got nil")
		}
	})
}
