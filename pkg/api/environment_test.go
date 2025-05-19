package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v45/github"
)

func TestEnvironmentFunctions(t *testing.T) {
	t.Run("GetEnvironmentPublicKey", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/repos/testorg/repo/environments/env/secrets/public-key" {
				t.Errorf("Unexpected URL path: %s", r.URL.Path)
				http.NotFound(w, r)
				return
			}

			response := struct {
				Key   string `json:"key"`
				KeyID string `json:"key_id"`
			}{
				Key:   valid32ByteKey,
				KeyID: "keyid",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		// Create a custom transport that redirects all requests to our test server
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

		encryption, err := client.GetEnvironmentPublicKey("testorg", "repo", "env")
		if err != nil {
			t.Fatalf("GetEnvironmentPublicKey returned error: %v", err)
		}

		if encryption.KeyID != "keyid" {
			t.Errorf("GetEnvironmentPublicKey returned KeyID %q, want %q", encryption.KeyID, "keyid")
		}

		if len(encryption.PublicKey) != 32 {
			t.Errorf("GetEnvironmentPublicKey returned PublicKey of length %d, want %d", len(encryption.PublicKey), 32)
		}
	})

	t.Run("ListEnvironmentSecrets", func(t *testing.T) {
		secretsResp := &github.Secrets{
			Secrets: []*github.Secret{{Name: "ENV_SECRET"}},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/repos/testorg/repo/environments/env/secrets" {
				t.Errorf("Unexpected URL path: %s", r.URL.Path)
				http.NotFound(w, r)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(secretsResp)
		}))
		defer server.Close()

		// Create a custom transport that redirects all requests to our test server
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

		secrets, err := client.ListEnvironmentSecrets("testorg", "repo", "env")
		if err != nil {
			t.Fatalf("ListEnvironmentSecrets returned error: %v", err)
		}

		if len(secrets) != 1 || secrets[0].Name != "ENV_SECRET" {
			t.Errorf("ListEnvironmentSecrets returned incorrect secrets: %+v", secrets)
		}
	})

	t.Run("CreateOrUpdateEnvironmentSecret", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/repos/testorg/repo/environments/env/secrets/public-key":
				response := struct {
					Key   string `json:"key"`
					KeyID string `json:"key_id"`
				}{
					Key:   valid32ByteKey,
					KeyID: "keyid",
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			case "/repos/testorg/repo/environments/env/secrets/SECRET":
				if r.Method != "PUT" {
					t.Errorf("Unexpected method: %s", r.Method)
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				w.WriteHeader(http.StatusNoContent)
			default:
				t.Errorf("Unexpected URL path: %s", r.URL.Path)
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		// Create a custom transport that redirects all requests to our test server
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

		secret := &github.EncryptedSecret{
			Name:           "SECRET",
			EncryptedValue: "encrypted-value",
		}

		err := client.CreateOrUpdateEnvironmentSecret("testorg", "repo", "env", secret)
		if err != nil {
			t.Fatalf("CreateOrUpdateEnvironmentSecret returned error: %v", err)
		}
	})

	t.Run("DeleteEnvironmentSecret", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/repos/testorg/repo/environments/env/secrets/SECRET" || r.Method != "DELETE" {
				t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
				http.NotFound(w, r)
				return
			}

			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		// Create a custom transport that redirects all requests to our test server
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

		err := client.DeleteEnvironmentSecret("testorg", "repo", "env", "SECRET")
		if err != nil {
			t.Fatalf("DeleteEnvironmentSecret returned error: %v", err)
		}
	})

	t.Run("GetEnvPublicKey", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/repos/testorg/repo/environments/env/secrets/public-key" {
				t.Errorf("Unexpected URL path: %s", r.URL.Path)
				http.NotFound(w, r)
				return
			}

			response := struct {
				Key   string `json:"key"`
				KeyID string `json:"key_id"`
			}{
				Key:   valid32ByteKey,
				KeyID: "env-keyid",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		// Create a custom transport that redirects all requests to our test server
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

		encryption, err := client.GetEnvPublicKey("testorg", "repo", "env")
		if err != nil {
			t.Fatalf("GetEnvPublicKey returned error: %v", err)
		}

		if *encryption.KeyID != "env-keyid" {
			t.Errorf("GetEnvPublicKey returned KeyID %q, want %q", *encryption.KeyID, "env-keyid")
		}

		decodedKey, err := base64.StdEncoding.DecodeString(*encryption.Key)
		if err != nil {
			t.Fatalf("Failed to decode key: %v", err)
		}
		if len(decodedKey) != 32 {
			t.Errorf("GetEnvPublicKey returned decoded key of length %d, want %d", len(decodedKey), 32)
		}
	})

	t.Run("ListEnvSecrets", func(t *testing.T) {
		secretsResp := &github.Secrets{
			Secrets: []*github.Secret{{Name: "ENV_SECRET_2"}},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/repos/testorg/repo/environments/env/secrets" {
				t.Errorf("Unexpected URL path: %s", r.URL.Path)
				http.NotFound(w, r)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(secretsResp)
		}))
		defer server.Close()

		// Create a custom transport that redirects all requests to our test server
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

		secrets, err := client.ListEnvSecrets("testorg", "repo", "env")
		if err != nil {
			t.Fatalf("ListEnvSecrets returned error: %v", err)
		}

		if len(secrets) != 1 || secrets[0].Name != "ENV_SECRET_2" {
			t.Errorf("ListEnvSecrets returned incorrect secrets: %+v", secrets)
		}
	})

	t.Run("CreateOrUpdateEnvSecret", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/repos/testorg/repo/environments/env/secrets/public-key":
				response := struct {
					Key   string `json:"key"`
					KeyID string `json:"key_id"`
				}{
					Key:   valid32ByteKey,
					KeyID: "keyid",
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			case "/repos/testorg/repo/environments/env/secrets/ENV_SECRET":
				if r.Method != "PUT" {
					t.Errorf("Unexpected method: %s", r.Method)
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}

				// Verify the request body
				var req struct {
					EncryptedValue string `json:"encrypted_value"`
					KeyID          string `json:"key_id"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Errorf("Failed to decode request body: %v", err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				if req.KeyID != "keyid" {
					t.Errorf("Expected KeyID %q, got %q", "keyid", req.KeyID)
				}

				w.WriteHeader(http.StatusNoContent)
			default:
				t.Errorf("Unexpected URL path: %s", r.URL.Path)
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		// Create a custom transport that redirects all requests to our test server
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

		secret := &github.EncryptedSecret{
			Name:           "ENV_SECRET",
			EncryptedValue: "encrypted-value",
		}

		err := client.CreateOrUpdateEnvSecret("testorg", "repo", "env", secret)
		if err != nil {
			t.Fatalf("CreateOrUpdateEnvSecret returned error: %v", err)
		}
	})

	t.Run("DeleteEnvSecret", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/repos/testorg/repo/environments/env/secrets/ENV_SECRET" || r.Method != "DELETE" {
				t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
				http.NotFound(w, r)
				return
			}

			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		// Create a custom transport that redirects all requests to our test server
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

		err := client.DeleteEnvSecret("testorg", "repo", "env", "ENV_SECRET")
		if err != nil {
			t.Fatalf("DeleteEnvSecret returned error: %v", err)
		}
	})

	t.Run("GetEnvironmentSecret", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/repos/testorg/repo/environments/env/secrets/ENV_SECRET" || r.Method != "GET" {
				t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
				http.NotFound(w, r)
				return
			}

			secret := &github.Secret{
				Name:      "ENV_SECRET",
				CreatedAt: github.Timestamp{},
				UpdatedAt: github.Timestamp{},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(secret)
		}))
		defer server.Close()

		// Create a custom transport that redirects all requests to our test server
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

		secret, err := client.GetEnvironmentSecret("testorg", "repo", "env", "ENV_SECRET")
		if err != nil {
			t.Fatalf("GetEnvironmentSecret returned error: %v", err)
		}

		if secret.Name != "ENV_SECRET" {
			t.Errorf("GetEnvironmentSecret returned name %q, want %q", secret.Name, "ENV_SECRET")
		}
	})
}
