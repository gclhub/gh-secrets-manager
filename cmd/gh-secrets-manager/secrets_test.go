package main

import (
	"encoding/json"
	"testing"
)

func TestSplitRepoInSecrets(t *testing.T) {
	testCases := []struct {
		name          string
		repo          string
		expectedOwner string
		expectedRepo  string
	}{
		{
			name:          "Valid repository",
			repo:          "owner/repo",
			expectedOwner: "owner",
			expectedRepo:  "repo",
		},
		{
			name:          "Missing slash",
			repo:          "ownerrepo",
			expectedOwner: "",
			expectedRepo:  "ownerrepo",
		},
		{
			name:          "Empty string",
			repo:          "",
			expectedOwner: "",
			expectedRepo:  "",
		},
		{
			name:          "Too many slashes",
			repo:          "owner/repo/extra",
			expectedOwner: "",
			expectedRepo:  "owner/repo/extra",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			owner, repo := splitRepo(tc.repo)

			if owner != tc.expectedOwner {
				t.Errorf("Expected owner %q, got %q", tc.expectedOwner, owner)
			}
			if repo != tc.expectedRepo {
				t.Errorf("Expected repo %q, got %q", tc.expectedRepo, repo)
			}
		})
	}
}

func TestOutputJSON(t *testing.T) {
	testCases := []struct {
		name          string
		data          interface{}
		expectError   bool
	}{
		{
			name:          "Valid data",
			data:          map[string]string{"key": "value"},
			expectError:   false,
		},
		{
			name:          "Slice data",
			data:          []string{"item1", "item2"},
			expectError:   false,
		},
		{
			name:          "Struct data",
			data:          struct{ Name string }{"Test"},
			expectError:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// We can't easily test the output itself since outputJSON writes to stdout
			// But we can verify it doesn't error with valid data
			err := outputJSON(tc.data)
			if tc.expectError && err == nil {
				t.Errorf("Expected an error, but got none")
			} else if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			// Additional check - make sure the data can be marshaled
			_, marshalErr := json.MarshalIndent(tc.data, "", "  ")
			if marshalErr != nil {
				t.Errorf("Data couldn't be marshaled to JSON: %v", marshalErr)
			}
		})
	}
}