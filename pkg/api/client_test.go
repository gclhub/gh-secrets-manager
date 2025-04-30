package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
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
		if r.URL.Path != path {
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

func TestListRepoSecrets(t *testing.T) {
	response := &github.Secrets{
		Secrets: []*github.Secret{
			{Name: "SECRET1"},
			{Name: "SECRET2"},
		},
	}

	server, client := setupTestServer(t, "/repos/owner/repo/actions/secrets", response)
	defer server.Close()

	result, err := client.ListRepoSecrets("owner", "repo")
	if err != nil {
		t.Fatalf("ListRepoSecrets returned error: %v", err)
	}

	if len(result) != len(response.Secrets) {
		t.Errorf("ListRepoSecrets returned %d secrets, want %d", len(result), len(response.Secrets))
	}

	for i, secret := range result {
		if secret.Name != response.Secrets[i].Name {
			t.Errorf("Secret %d: got name %q, want %q", i, secret.Name, response.Secrets[i].Name)
		}
	}
}

func TestListOrgVariables(t *testing.T) {
	expectedVars := []*Variable{
		{Name: "VAR1", Value: "value1"},
		{Name: "VAR2", Value: "value2"},
	}

	response := struct {
		Variables []*Variable `json:"variables"`
	}{
		Variables: expectedVars,
	}

	server, client := setupTestServer(t, "/orgs/testorg/actions/variables", response)
	defer server.Close()

	variables, err := client.ListOrgVariables("testorg")
	if err != nil {
		t.Fatalf("ListOrgVariables returned error: %v", err)
	}

	if len(variables) != len(expectedVars) {
		t.Errorf("ListOrgVariables returned %d variables, want %d", len(variables), len(expectedVars))
	}

	for i, variable := range variables {
		if variable.Name != expectedVars[i].Name || variable.Value != expectedVars[i].Value {
			t.Errorf("Variable %d: got %v, want %v", i, variable, expectedVars[i])
		}
	}
}

func TestListRepoVariables(t *testing.T) {
	expectedVars := []*Variable{
		{Name: "VAR1", Value: "value1"},
		{Name: "VAR2", Value: "value2"},
	}

	response := struct {
		Variables []*Variable `json:"variables"`
	}{
		Variables: expectedVars,
	}

	server, client := setupTestServer(t, "/repos/owner/repo/actions/variables", response)
	defer server.Close()

	variables, err := client.ListRepoVariables("owner", "repo")
	if err != nil {
		t.Fatalf("ListRepoVariables returned error: %v", err)
	}

	if len(variables) != len(expectedVars) {
		t.Errorf("ListRepoVariables returned %d variables, want %d", len(variables), len(expectedVars))
	}

	for i, variable := range variables {
		if variable.Name != expectedVars[i].Name || variable.Value != expectedVars[i].Value {
			t.Errorf("Variable %d: got %v, want %v", i, variable, expectedVars[i])
		}
	}
}

func TestCreateOrUpdateOrgVariable(t *testing.T) {
	variable := &Variable{
		Name:  "TEST_VAR",
		Value: "test_value",
	}

	server, client := setupTestServer(t, "/orgs/testorg/actions/variables/TEST_VAR", nil)
	defer server.Close()

	err := client.CreateOrUpdateOrgVariable("testorg", variable)
	if err != nil {
		t.Errorf("CreateOrUpdateOrgVariable returned error: %v", err)
	}
}

func TestCreateOrUpdateRepoVariable(t *testing.T) {
	variable := &Variable{
		Name:  "TEST_VAR",
		Value: "test_value",
	}

	server, client := setupTestServer(t, "/repos/owner/repo/actions/variables/TEST_VAR", nil)
	defer server.Close()

	err := client.CreateOrUpdateRepoVariable("owner", "repo", variable)
	if err != nil {
		t.Errorf("CreateOrUpdateRepoVariable returned error: %v", err)
	}
}

func TestDeleteOrgVariable(t *testing.T) {
	server, client := setupTestServer(t, "/orgs/testorg/actions/variables/TEST_VAR", nil)
	defer server.Close()

	err := client.DeleteOrgVariable("testorg", "TEST_VAR")
	if err != nil {
		t.Errorf("DeleteOrgVariable returned error: %v", err)
	}
}

func TestDeleteRepoVariable(t *testing.T) {
	server, client := setupTestServer(t, "/repos/owner/repo/actions/variables/TEST_VAR", nil)
	defer server.Close()

	err := client.DeleteRepoVariable("owner", "repo", "TEST_VAR")
	if err != nil {
		t.Errorf("DeleteRepoVariable returned error: %v", err)
	}
}

func TestListRepositoriesByProperty(t *testing.T) {
	repos := []*github.Repository{
		{
			Name:     github.String("repo1"),
			Language: github.String("go"),
		},
		{
			Name:     github.String("repo2"),
			Language: github.String("python"),
		},
		{
			Name:     github.String("repo3"),
			Language: github.String("go"),
		},
	}

	server, client := setupTestServer(t, "/orgs/testorg/repos", repos)
	defer server.Close()

	matchingRepos, err := client.ListRepositoriesByProperty("testorg", "language", "go")
	if err != nil {
		t.Fatalf("ListRepositoriesByProperty returned error: %v", err)
	}

	if len(matchingRepos) != 2 {
		t.Errorf("ListRepositoriesByProperty returned %d repos, want 2", len(matchingRepos))
	}

	for _, repo := range matchingRepos {
		if repo.GetLanguage() != "go" {
			t.Errorf("Repository %s has language %s, want 'go'", repo.GetName(), repo.GetLanguage())
		}
	}
}
