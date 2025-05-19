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
	"encoding/base64"
)

// Shared struct and variable for public key responses in tests
// (Must be at package scope for all handlers to access)
type pk struct {
	Key   string `json:"key"`
	KeyID string `json:"key_id"`
}
var valid32ByteKey = base64.StdEncoding.EncodeToString(make([]byte, 32))

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

	// Create a custom transport that redirects all api.github.com requests to our test server
	baseURL, _ := url.Parse(server.URL + "/")
	httpClient := newTestClient(baseURL)
	
	// Create a GitHub client that uses our custom transport
	ghClient := github.NewClient(httpClient)
	ghClient.BaseURL = baseURL
	
	client := &Client{
		github: ghClient,
		ctx:    context.Background(),
		opts:   &ClientOptions{AuthMethod: AuthMethodPAT}, // Use PAT auth for testing
	}
	client.github.BaseURL, _ = url.Parse(server.URL + "/")

	return server, client
}

func setupMultiHandlerTestServer(t *testing.T, handlers map[string]http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[TEST HANDLER] Incoming path: %s\n", r.URL.Path)
		for prefix, handler := range handlers {
			if r.URL.Path == prefix || strings.HasPrefix(r.URL.Path, prefix+"/") {
				fmt.Printf("[TEST HANDLER] Matched handler for prefix: %s\n", prefix)
				handler(w, r)
				return
			}
		}
		if strings.HasSuffix(r.URL.Path, "/public-key") {
			fmt.Printf("[TEST HANDLER] Catch-all for /public-key: %s\n", r.URL.Path)
			json.NewEncoder(w).Encode(pk{Key: valid32ByteKey, KeyID: "keyid"})
			return
		}
		fmt.Printf("[TEST HANDLER] No handler matched for path: %s\n", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}))

	// Create a custom transport that redirects all api.github.com requests to our test server
	baseURL, _ := url.Parse(server.URL + "/")
	httpClient := newTestClient(baseURL)
	
	// Create a GitHub client that uses our custom transport
	ghClient := github.NewClient(httpClient)
	ghClient.BaseURL = baseURL
	
	client := &Client{
		github: ghClient,
		ctx:    context.Background(),
		opts:   &ClientOptions{AuthMethod: AuthMethodPAT},
	}

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

		repos := map[string]*github.Repository{
			"repo1": {Name: github.String("repo1"), FullName: github.String("testorg/repo1")},
			"repo3": {Name: github.String("repo3"), FullName: github.String("testorg/repo3")},
		}

		handlers := map[string]http.HandlerFunc{
			"/orgs/testorg/properties/values": func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(propertyResponse)
			},
			"/repos/testorg/repo1": func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(repos["repo1"])
			},
			"/repos/testorg/repo3": func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(repos["repo3"])
			},
		}

		server, client := setupMultiHandlerTestServer(t, handlers)
		defer server.Close()

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

		repos := map[string]*github.Repository{
			"repo1": {Name: github.String("repo1"), FullName: github.String("testorg/repo1")},
			"repo3": {Name: github.String("repo3"), FullName: github.String("testorg/repo3")},
		}

		handlers := map[string]http.HandlerFunc{
			"/orgs/testorg/properties/values": func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(responses[pageNum])
				if pageNum < len(responses)-1 {
					pageNum++
				}
			},
			"/repos/testorg/repo1": func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(repos["repo1"])
			},
			"/repos/testorg/repo3": func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(repos["repo3"])
			},
		}

		server, client := setupMultiHandlerTestServer(t, handlers)
		defer server.Close()

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
		handlers := map[string]http.HandlerFunc{
			"/orgs/testorg/properties/values": func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
		}
		server, client := setupMultiHandlerTestServer(t, handlers)
		defer server.Close()

		_, err := client.ListRepositoriesByProperty("testorg", "team", "backend")
		if err == nil {
			t.Error("Expected error but got nil")
		}
	})

	t.Run("missing_org", func(t *testing.T) {
		handlers := map[string]http.HandlerFunc{
			"/orgs/testorg/properties/values": func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		}
		server, client := setupMultiHandlerTestServer(t, handlers)
		defer server.Close()

		_, err := client.ListRepositoriesByProperty("", "team", "backend")
		if err == nil {
			t.Error("Expected error but got nil")
		} else if !strings.Contains(err.Error(), "organization name") {
			t.Errorf("Expected error about missing organization name, got: %v", err)
		}
	})

	t.Run("missing_property_name", func(t *testing.T) {
		handlers := map[string]http.HandlerFunc{
			"/orgs/testorg/properties/values": func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		}
		server, client := setupMultiHandlerTestServer(t, handlers)
		defer server.Close()

		_, err := client.ListRepositoriesByProperty("testorg", "", "backend")
		if err == nil {
			t.Error("Expected error but got nil")
		} else if !strings.Contains(err.Error(), "property_name") {
			t.Errorf("Expected error about missing property_name, got: %v", err)
		}
	})

	t.Run("missing_property_value", func(t *testing.T) {
		handlers := map[string]http.HandlerFunc{
			"/orgs/testorg/properties/values": func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		}
		server, client := setupMultiHandlerTestServer(t, handlers)
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

		handlers := map[string]http.HandlerFunc{
			"/orgs/testorg/properties/values": func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(propertyResponse)
			},
			"/repos/testorg/": func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
		}
		server, client := setupMultiHandlerTestServer(t, handlers)
		defer server.Close()

		_, err := client.ListRepositoriesByProperty("testorg", "team", "backend")
		if err == nil {
			t.Error("Expected error but got nil")
		}
	})
}

func setupComprehensiveHandler(t *testing.T) (*httptest.Server, *Client) {
	variable := &Variable{Name: "VAR", Value: "value"}
	secretsResp := &github.Secrets{Secrets: []*github.Secret{{Name: "SECRET"}}}
	varsResp := struct {
		Variables []*Variable `json:"variables"`
	}{Variables: []*Variable{variable}}

	handlers := map[string]http.HandlerFunc{
		// All org secrets endpoints
		"/orgs/testorg/actions/secrets": func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(secretsResp)
		},
		"/orgs/testorg/actions/secrets/SECRET": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		},
		
		// Specific public key endpoints
		"/orgs/testorg/actions/secrets/public-key": func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("[TEST HANDLER] Processing org public key request: %s %s\n", r.Method, r.URL.Path)
			response := struct {
				Key   string `json:"key"`
				KeyID string `json:"key_id"`
			}{
				Key:   valid32ByteKey,
				KeyID: "keyid",
			}
			fmt.Printf("[TEST HANDLER] Sending response: %+v\n", response)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		},
		"/repos/testorg/repo/actions/secrets/public-key": func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("[TEST HANDLER] Processing repo public key request: %s %s\n", r.Method, r.URL.Path)
			response := struct {
				Key   string `json:"key"`
				KeyID string `json:"key_id"`
			}{
				Key:   valid32ByteKey,
				KeyID: "keyid",
			}
			fmt.Printf("[TEST HANDLER] Sending response: %+v\n", response)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		},
		"/orgs/testorg/dependabot/secrets/public-key": func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("[TEST HANDLER] Processing org dependabot public key request: %s %s\n", r.Method, r.URL.Path)
			response := struct {
				Key   string `json:"key"`
				KeyID string `json:"key_id"`
			}{
				Key:   valid32ByteKey,
				KeyID: "keyid",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		},
		"/repos/testorg/repo/dependabot/secrets/public-key": func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("[TEST HANDLER] Processing repo dependabot public key request: %s %s\n", r.Method, r.URL.Path)
			response := struct {
				Key   string `json:"key"`
				KeyID string `json:"key_id"`
			}{
				Key:   valid32ByteKey,
				KeyID: "keyid",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		},
		"/repos/testorg/repo/environments/env/secrets/public-key": func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("[TEST HANDLER] Processing env public key request: %s %s\n", r.Method, r.URL.Path)
			response := struct {
				Key   string `json:"key"`
				KeyID string `json:"key_id"`
			}{
				Key:   valid32ByteKey,
				KeyID: "keyid",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		},
		
		// All repo secrets endpoints
		"/repos/testorg/repo/actions/secrets": func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(secretsResp)
		},
		"/repos/testorg/repo/actions/secrets/SECRET": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		},
		
		// All variables endpoints
		"/orgs/testorg/actions/variables": func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(varsResp)
		},
		"/orgs/testorg/actions/variables/VAR": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		},
		"/repos/testorg/repo/actions/variables": func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(varsResp)
		},
		"/repos/testorg/repo/actions/variables/VAR": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		},
		
		// All dependabot endpoints
		"/orgs/testorg/dependabot/secrets": func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(secretsResp)
		},
		"/orgs/testorg/dependabot/secrets/SECRET": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		},
		"/repos/testorg/repo/dependabot/secrets": func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(secretsResp)
		},
		"/repos/testorg/repo/dependabot/secrets/SECRET": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		},
		
		// All env variables/secrets endpoints
		"/repos/testorg/repo/environments/env/variables": func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(varsResp)
		},
		"/repos/testorg/repo/environments/env/variables/VAR": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[TEST SERVER] Received request: %s %s\n", r.Method, r.URL.Path)
		
		// First, check for exact path matches (important for public key endpoints)
		if handler, exists := handlers[r.URL.Path]; exists {
			fmt.Printf("[TEST SERVER] Matched exact handler for path: %s\n", r.URL.Path)
			handler(w, r)
			return
		}
		
		// Then check for prefix matches
		for prefix, handler := range handlers {
			if strings.HasPrefix(r.URL.Path, prefix+"/") {
				fmt.Printf("[TEST SERVER] Matched handler for prefix: %s\n", prefix)
				handler(w, r)
				return
			}
		}
		
		// Catch-all for any public-key endpoint
		if strings.Contains(r.URL.Path, "/public-key") {
			fmt.Printf("[TEST SERVER] Catch-all for public-key: %s\n", r.URL.Path)
			response := struct {
				Key   string `json:"key"`
				KeyID string `json:"key_id"`
			}{
				Key:   valid32ByteKey,
				KeyID: "keyid",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
		
		fmt.Printf("[TEST SERVER] No handler matched for path: %s\n", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}))

	// Create a custom transport that redirects all api.github.com requests to our test server
	baseURL, _ := url.Parse(server.URL + "/")
	httpClient := newTestClient(baseURL)
	
	// Create a GitHub client that uses our custom transport
	ghClient := github.NewClient(httpClient)
	ghClient.BaseURL = baseURL
	
	client := &Client{
		github: ghClient,
		ctx:    context.Background(),
		opts:   &ClientOptions{AuthMethod: AuthMethodPAT},
	}

	return server, client
}

func TestClient_SecretsAndVariablesCRUD(t *testing.T) {
	// Table-driven tests for all CRUD operations and error cases for org/repo/env secrets, variables, and dependabot
	type testCase struct {
		name       string
		method     string
		args       []any
		setup      func(*testing.T) (*httptest.Server, *Client)
		expectErr  bool
		errorMatch string
	}

	secret := &github.EncryptedSecret{Name: "SECRET", EncryptedValue: "encrypted"}
	variable := &Variable{Name: "VAR", Value: "value"}

	tests := []testCase{
		// Org secrets
		{"ListOrgSecrets OK", "ListOrgSecrets", []any{"testorg"}, setupComprehensiveHandler, false, ""},
		{"CreateOrUpdateOrgSecret OK", "CreateOrUpdateOrgSecret", []any{"testorg", secret}, setupComprehensiveHandler, false, ""},
		{"DeleteOrgSecret OK", "DeleteOrgSecret", []any{"testorg", "SECRET"}, setupComprehensiveHandler, false, ""},
		// Repo secrets
		{"ListRepoSecrets OK", "ListRepoSecrets", []any{"testorg", "repo"}, setupComprehensiveHandler, false, ""},
		{"CreateOrUpdateRepoSecret OK", "CreateOrUpdateRepoSecret", []any{"testorg", "repo", secret}, setupComprehensiveHandler, false, ""},
		{"DeleteRepoSecret OK", "DeleteRepoSecret", []any{"testorg", "repo", "SECRET"}, setupComprehensiveHandler, false, ""},
		// Org variables
		{"ListOrgVariables OK", "ListOrgVariables", []any{"testorg"}, setupComprehensiveHandler, false, ""},
		{"CreateOrUpdateOrgVariable OK", "CreateOrUpdateOrgVariable", []any{"testorg", variable}, setupComprehensiveHandler, false, ""},
		{"DeleteOrgVariable OK", "DeleteOrgVariable", []any{"testorg", "VAR"}, setupComprehensiveHandler, false, ""},
		// Repo variables
		{"ListRepoVariables OK", "ListRepoVariables", []any{"testorg", "repo"}, setupComprehensiveHandler, false, ""},
		{"CreateOrUpdateRepoVariable OK", "CreateOrUpdateRepoVariable", []any{"testorg", "repo", variable}, setupComprehensiveHandler, false, ""},
		{"DeleteRepoVariable OK", "DeleteRepoVariable", []any{"testorg", "repo", "VAR"}, setupComprehensiveHandler, false, ""},
		// Dependabot org/repo secrets
		{"ListOrgDependabotSecrets OK", "ListOrgDependabotSecrets", []any{"testorg"}, setupComprehensiveHandler, false, ""},
		{"CreateOrUpdateOrgDependabotSecret OK", "CreateOrUpdateOrgDependabotSecret", []any{"testorg", secret}, setupComprehensiveHandler, false, ""},
		{"DeleteOrgDependabotSecret OK", "DeleteOrgDependabotSecret", []any{"testorg", "SECRET"}, setupComprehensiveHandler, false, ""},
		{"ListRepoDependabotSecrets OK", "ListRepoDependabotSecrets", []any{"testorg", "repo"}, setupComprehensiveHandler, false, ""},
		{"CreateOrUpdateRepoDependabotSecret OK", "CreateOrUpdateRepoDependabotSecret", []any{"testorg", "repo", secret}, setupComprehensiveHandler, false, ""},
		{"DeleteRepoDependabotSecret OK", "DeleteRepoDependabotSecret", []any{"testorg", "repo", "SECRET"}, setupComprehensiveHandler, false, ""},
		// Environment secrets/variables (wrappers)
		{"ListEnvVariables OK", "ListEnvVariables", []any{"testorg", "repo", "env"}, setupComprehensiveHandler, false, ""},
		{"CreateOrUpdateEnvVariable OK", "CreateOrUpdateEnvVariable", []any{"testorg", "repo", "env", variable}, setupComprehensiveHandler, false, ""},
		{"DeleteEnvVariable OK", "DeleteEnvVariable", []any{"testorg", "repo", "env", "VAR"}, setupComprehensiveHandler, false, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server, client := tc.setup(t)
			defer server.Close()
			var err error
			switch tc.method {
			case "ListOrgSecrets":
				_, err = client.ListOrgSecrets(tc.args[0].(string))
			case "CreateOrUpdateOrgSecret":
				err = client.CreateOrUpdateOrgSecret(tc.args[0].(string), tc.args[1].(*github.EncryptedSecret))
			case "DeleteOrgSecret":
				err = client.DeleteOrgSecret(tc.args[0].(string), tc.args[1].(string))
			case "ListRepoSecrets":
				_, err = client.ListRepoSecrets(tc.args[0].(string), tc.args[1].(string))
			case "CreateOrUpdateRepoSecret":
				err = client.CreateOrUpdateRepoSecret(tc.args[0].(string), tc.args[1].(string), tc.args[2].(*github.EncryptedSecret))
			case "DeleteRepoSecret":
				err = client.DeleteRepoSecret(tc.args[0].(string), tc.args[1].(string), tc.args[2].(string))
			case "ListOrgVariables":
				_, err = client.ListOrgVariables(tc.args[0].(string))
			case "CreateOrUpdateOrgVariable":
				err = client.CreateOrUpdateOrgVariable(tc.args[0].(string), tc.args[1].(*Variable))
			case "DeleteOrgVariable":
				err = client.DeleteOrgVariable(tc.args[0].(string), tc.args[1].(string))
			case "ListRepoVariables":
				_, err = client.ListRepoVariables(tc.args[0].(string), tc.args[1].(string))
			case "CreateOrUpdateRepoVariable":
				err = client.CreateOrUpdateRepoVariable(tc.args[0].(string), tc.args[1].(string), tc.args[2].(*Variable))
			case "DeleteRepoVariable":
				err = client.DeleteRepoVariable(tc.args[0].(string), tc.args[1].(string), tc.args[2].(string))
			case "ListOrgDependabotSecrets":
				_, err = client.ListOrgDependabotSecrets(tc.args[0].(string))
			case "CreateOrUpdateOrgDependabotSecret":
				err = client.CreateOrUpdateOrgDependabotSecret(tc.args[0].(string), tc.args[1].(*github.EncryptedSecret))
			case "DeleteOrgDependabotSecret":
				err = client.DeleteOrgDependabotSecret(tc.args[0].(string), tc.args[1].(string))
			case "ListRepoDependabotSecrets":
				_, err = client.ListRepoDependabotSecrets(tc.args[0].(string), tc.args[1].(string))
			case "CreateOrUpdateRepoDependabotSecret":
				err = client.CreateOrUpdateRepoDependabotSecret(tc.args[0].(string), tc.args[1].(string), tc.args[2].(*github.EncryptedSecret))
			case "DeleteRepoDependabotSecret":
				err = client.DeleteRepoDependabotSecret(tc.args[0].(string), tc.args[1].(string), tc.args[2].(string))
			case "ListEnvVariables":
				_, err = client.ListEnvVariables(tc.args[0].(string), tc.args[1].(string), tc.args[2].(string))
			case "CreateOrUpdateEnvVariable":
				err = client.CreateOrUpdateEnvVariable(tc.args[0].(string), tc.args[1].(string), tc.args[2].(string), tc.args[3].(*Variable))
			case "DeleteEnvVariable":
				err = client.DeleteEnvVariable(tc.args[0].(string), tc.args[1].(string), tc.args[2].(string), tc.args[3].(string))
			}
			if tc.expectErr {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else if tc.errorMatch != "" && !strings.Contains(err.Error(), tc.errorMatch) {
					t.Errorf("Expected error containing %q, got %v", tc.errorMatch, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// --- MOCK DEPENDABOT SERVICE FOR TESTS ---
// (No longer needed after refactor)
// type mockDependabotService struct {
// 	t        *testing.T
// 	secrets  map[string]*github.EncryptedSecret
// 	variables map[string]*Variable
// }

// func newMockDependabotService(t *testing.T) *mockDependabotService {
// 	return &mockDependabotService{
// 		t        t,
// 		secrets:  make(map[string]*github.EncryptedSecret),
// 		variables: make(map[string]*Variable),
// 	}
// }

// func (m *mockDependabotService) ListOrgSecrets(org string) ([]*github.EncryptedSecret, error) {
// 	m.t.Logf("ListOrgSecrets called with org: %s", org)
// 	var result []*github.EncryptedSecret
// 	for _, secret := range m.secrets {
// 		result = append(result, secret)
// 	}
// 	return result, nil
// }

// func (m *mockDependabotService) CreateOrUpdateOrgSecret(org string, secret *github.EncryptedSecret) error {
// 	m.t.Logf("CreateOrUpdateOrgSecret called with org: %s, secret: %+v", org, secret)
// 	m.secrets[secret.Name] = secret
// 	return nil
// }

// func (m *mockDependabotService) DeleteOrgSecret(org, secretName string) error {
// 	m.t.Logf("DeleteOrgSecret called with org: %s, secretName: %s", org, secretName)
// 	delete(m.secrets, secretName)
// 	return nil
// }

// func (m *mockDependabotService) ListRepoSecrets(org, repo string) ([]*github.EncryptedSecret, error) {
// 	m.t.Logf("ListRepoSecrets called with org: %s, repo: %s", org, repo)
// 	var result []*github.EncryptedSecret
// 	for _, secret := range m.secrets {
// 		result = append(result, secret)
// 	}
// 	return result, nil
// }

// func (m *mockDependabotService) CreateOrUpdateRepoSecret(org, repo string, secret *github.EncryptedSecret) error {
// 	m.t.Logf("CreateOrUpdateRepoSecret called with org: %s, repo: %s, secret: %+v", org, repo, secret)
// 	m.secrets[secret.Name] = secret
// 	return nil
// }

// func (m *mockDependabotService) DeleteRepoSecret(org, repo, secretName string) error {
// 	m.t.Logf("DeleteRepoSecret called with org: %s, repo: %s, secretName: %s", org, repo, secretName)
// 	delete(m.secrets, secretName)
// 	return nil
// }

// func (m *mockDependabotService) ListOrgVariables(org string) ([]*Variable, error) {
// 	m.t.Logf("ListOrgVariables called with org: %s", org)
// 	var result []*Variable
// 	for _, variable := range m.variables {
// 		result = append(result, variable)
// 	}
// 	return result, nil
// }

// func (m *mockDependabotService) CreateOrUpdateOrgVariable(org string, variable *Variable) error {
// 	m.t.Logf("CreateOrUpdateOrgVariable called with org: %s, variable: %+v", org, variable)
// 	m.variables[variable.Name] = variable
// 	return nil
// }

// func (m *mockDependabotService) DeleteOrgVariable(org, variableName string) error {
// 	m.t.Logf("DeleteOrgVariable called with org: %s, variableName: %s", org, variableName)
// 	delete(m.variables, variableName)
// 	return nil
// }

// func (m *mockDependabotService) ListRepoVariables(org, repo string) ([]*Variable, error) {
// 	m.t.Logf("ListRepoVariables called with org: %s, repo: %s", org, repo)
// 	var result []*Variable
// 	for _, variable := range m.variables {
// 		result = append(result, variable)
// 	}
// 	return result, nil
// }

// func (m *mockDependabotService) CreateOrUpdateRepoVariable(org, repo string, variable *Variable) error {
// 	m.t.Logf("CreateOrUpdateRepoVariable called with org: %s, repo: %s, variable: %+v", org, repo, variable)
// 	m.variables[variable.Name] = variable
// 	return nil
// }

// func (m *mockDependabotService) DeleteRepoVariable(org, repo, variableName) error {
// 	m.t.Logf("DeleteRepoVariable called with org: %s, repo: %s, variableName: %s", org, repo, variableName)
// 	delete(m.variables, variableName)
// 	return nil
// }
