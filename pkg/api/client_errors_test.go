package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v45/github"
)

func TestListOrgSecrets_Error(t *testing.T) {
	// Setup test server to return an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()
	
	// Create client pointing to our error server
	baseURL, _ := url.Parse(server.URL + "/")
	httpClient := newTestClient(baseURL)
	
	client := &Client{
		github: github.NewClient(httpClient),
		ctx:    nil,
		opts:   &ClientOptions{AuthMethod: AuthMethodPAT},
	}
	
	// This should return an error
	_, err := client.ListOrgSecrets("testorg")
	if err == nil {
		t.Error("ListOrgSecrets should have returned an error")
	}
}

func TestGetOrgDependabotPublicKey_Error(t *testing.T) {
	// Setup test server to return an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()
	
	// Create client pointing to our error server
	baseURL, _ := url.Parse(server.URL + "/")
	httpClient := newTestClient(baseURL)
	
	client := &Client{
		github: github.NewClient(httpClient),
		ctx:    nil,
		opts:   &ClientOptions{AuthMethod: AuthMethodPAT},
	}
	
	// This should return an error
	_, err := client.GetOrgDependabotPublicKey("testorg")
	if err == nil {
		t.Error("GetOrgDependabotPublicKey should have returned an error")
	}
}

func TestGetRepoDependabotPublicKey_Error(t *testing.T) {
	// Setup test server to return an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()
	
	// Create client pointing to our error server
	baseURL, _ := url.Parse(server.URL + "/")
	httpClient := newTestClient(baseURL)
	
	client := &Client{
		github: github.NewClient(httpClient),
		ctx:    nil,
		opts:   &ClientOptions{AuthMethod: AuthMethodPAT},
	}
	
	// This should return an error
	_, err := client.GetRepoDependabotPublicKey("testorg", "testrepo")
	if err == nil {
		t.Error("GetRepoDependabotPublicKey should have returned an error")
	}
}

func TestListEnvironmentVariables_Error(t *testing.T) {
	// Setup test server to return an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()
	
	// Create client pointing to our error server
	baseURL, _ := url.Parse(server.URL + "/")
	httpClient := newTestClient(baseURL)
	
	client := &Client{
		github: github.NewClient(httpClient),
		ctx:    nil,
		opts:   &ClientOptions{AuthMethod: AuthMethodPAT},
	}
	
	// This should return an error
	_, err := client.ListEnvironmentVariables("testorg", "testrepo", "testenv")
	if err == nil {
		t.Error("ListEnvironmentVariables should have returned an error")
	}
}

func TestCreateOrUpdateEnvironmentVariable_Error(t *testing.T) {
	// Setup test server to return an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()
	
	// Create client pointing to our error server
	baseURL, _ := url.Parse(server.URL + "/")
	httpClient := newTestClient(baseURL)
	
	client := &Client{
		github: github.NewClient(httpClient),
		ctx:    nil,
		opts:   &ClientOptions{AuthMethod: AuthMethodPAT},
	}
	
	// This should return an error
	variable := &Variable{Name: "TEST_VAR", Value: "test-value"}
	err := client.CreateOrUpdateEnvironmentVariable("testorg", "testrepo", "testenv", variable)
	if err == nil {
		t.Error("CreateOrUpdateEnvironmentVariable should have returned an error")
	}
}

func TestDeleteEnvironmentVariable_Error(t *testing.T) {
	// Setup test server to return an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()
	
	// Create client pointing to our error server
	baseURL, _ := url.Parse(server.URL + "/")
	httpClient := newTestClient(baseURL)
	
	client := &Client{
		github: github.NewClient(httpClient),
		ctx:    nil,
		opts:   &ClientOptions{AuthMethod: AuthMethodPAT},
	}
	
	// This should return an error
	err := client.DeleteEnvironmentVariable("testorg", "testrepo", "testenv", "TEST_VAR")
	if err == nil {
		t.Error("DeleteEnvironmentVariable should have returned an error")
	}
}
