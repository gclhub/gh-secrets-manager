package main

import (
	"testing"
)

func TestValidateConfigKey(t *testing.T) {
	testCases := []struct {
		name     string
		key      string
		expectError bool
	}{
		{
			name:     "Valid key - auth-server",
			key:      "auth-server",
			expectError: false,
		},
		{
			name:     "Valid key - app-id",
			key:      "app-id",
			expectError: false,
		},
		{
			name:     "Valid key - installation-id",
			key:      "installation-id",
			expectError: false,
		},
		{
			name:     "Invalid key",
			key:      "invalid-key",
			expectError: true,
		},
		{
			name:     "Empty key",
			key:      "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateConfigKey(tc.key)
			if tc.expectError && err == nil {
				t.Errorf("Expected error for key %q, got nil", tc.key)
			} else if !tc.expectError && err != nil {
				t.Errorf("Expected no error for key %q, got %v", tc.key, err)
			}
		})
	}
}

func TestValidateIntegerValue(t *testing.T) {
	testCases := []struct {
		name        string
		value       string
		key         string
		expectError bool
	}{
		{
			name:        "Valid integer",
			value:       "12345",
			key:         "app-id",
			expectError: false,
		},
		{
			name:        "Negative integer",
			value:       "-123",
			key:         "app-id",
			expectError: true,
		},
		{
			name:        "Zero",
			value:       "0",
			key:         "app-id",
			expectError: true,
		},
		{
			name:        "Invalid - contains letters",
			value:       "123abc",
			key:         "app-id",
			expectError: true,
		},
		{
			name:        "Invalid - float",
			value:       "123.45",
			key:         "app-id",
			expectError: true,
		},
		{
			name:        "Invalid - empty",
			value:       "",
			key:         "app-id",
			expectError: true,
		},
		{
			name:        "Invalid - whitespace",
			value:       "  ",
			key:         "app-id",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := validateIntegerValue(tc.key, tc.value)
			if tc.expectError && err == nil {
				t.Errorf("Expected error for value %q, got nil", tc.value)
			} else if !tc.expectError && err != nil {
				t.Errorf("Expected no error for value %q, got %v", tc.value, err)
			}
		})
	}
}

func TestIsLocalhostAddress(t *testing.T) {
	testCases := []struct {
		name     string
		address  string
		expected bool
	}{
		{
			name:     "localhost",
			address:  "localhost",
			expected: true,
		},
		{
			name:     "127.0.0.1",
			address:  "127.0.0.1",
			expected: true,
		},
		{
			name:     "::1",
			address:  "::1",
			expected: true,
		},
		{
			name:     "localhost with port",
			address:  "localhost:8080",
			expected: true,
		},
		{
			name:     "127.0.0.1 with port",
			address:  "127.0.0.1:8080",
			expected: true,
		},
		{
			name:     "::1 with port",
			address:  "[::1]:8080",
			expected: true,
		},
		{
			name:     "external IP",
			address:  "192.168.1.1",
			expected: false,
		},
		{
			name:     "external domain",
			address:  "example.com",
			expected: false,
		},
		{
			name:     "empty string",
			address:  "",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isLocalhostAddress(tc.address)
			if result != tc.expected {
				t.Errorf("Expected isLocalhostAddress(%q) to return %v, got %v", tc.address, tc.expected, result)
			}
		})
	}
}

func TestValidateURLValue(t *testing.T) {
	testCases := []struct {
		name        string
		value       string
		expectError bool
	}{
		{
			name:        "Valid HTTPS URL",
			value:       "https://example.com",
			expectError: false,
		},
		{
			name:        "Valid URL with path",
			value:       "https://example.com/api/v1",
			expectError: false,
		},
		{
			name:        "Valid URL with port",
			value:       "https://example.com:8080",
			expectError: false,
		},
		{
			name:        "Valid localhost HTTP URL",
			value:       "http://localhost:3000",
			expectError: false,
		},
		{
			name:        "Invalid - HTTP for non-localhost",
			value:       "http://example.com",
			expectError: true,
		},
		{
			name:        "Invalid - missing scheme",
			value:       "example.com",
			expectError: true,
		},
		{
			name:        "Invalid - unsupported scheme",
			value:       "ftp://example.com",
			expectError: true,
		},
		{
			name:        "Invalid - empty",
			value:       "",
			expectError: true,
		},
		{
			name:        "Invalid - malformed URL",
			value:       "http://",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateURLValue(tc.value)
			if tc.expectError && err == nil {
				t.Errorf("Expected error for value %q, got nil", tc.value)
			} else if !tc.expectError && err != nil {
				t.Errorf("Expected no error for value %q, got %v", tc.value, err)
			}
		})
	}
}

func TestValidateConfigValue(t *testing.T) {
	testCases := []struct {
		name        string
		key         string
		value       string
		expectError bool
	}{
		{
			name:        "Valid auth-server URL",
			key:         "auth-server",
			value:       "https://example.com",
			expectError: false,
		},
		{
			name:        "Invalid auth-server URL",
			key:         "auth-server",
			value:       "not-a-url",
			expectError: true,
		},
		{
			name:        "Valid app-id integer",
			key:         "app-id",
			value:       "12345",
			expectError: false,
		},
		{
			name:        "Invalid app-id non-integer",
			key:         "app-id",
			value:       "abc",
			expectError: true,
		},
		{
			name:        "Valid installation-id integer",
			key:         "installation-id",
			value:       "67890",
			expectError: false,
		},
		{
			name:        "Invalid installation-id non-integer",
			key:         "installation-id",
			value:       "xyz",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateConfigValue(tc.key, tc.value)
			if tc.expectError && err == nil {
				t.Errorf("Expected error for key %q and value %q, got nil", tc.key, tc.value)
			} else if !tc.expectError && err != nil {
				t.Errorf("Expected no error for key %q and value %q, got %v", tc.key, tc.value, err)
			}
		})
	}
}