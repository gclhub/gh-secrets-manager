package config

import (
	"os"
	"testing"
)

// TestIsGitHubAppConfigured checks if the GitHub App configuration detection works correctly
func TestIsGitHubAppConfigured(t *testing.T) {
	testCases := []struct {
		name     string
		config   Config
		expected bool
	}{
		{
			name:     "Empty config",
			config:   Config{},
			expected: false,
		},
		{
			name: "Only AuthServer configured",
			config: Config{
				AuthServer: "https://example.com",
			},
			expected: false,
		},
		{
			name: "Only AppID configured",
			config: Config{
				AppID: 12345,
			},
			expected: false,
		},
		{
			name: "Only InstallationID configured",
			config: Config{
				InstallationID: 67890,
			},
			expected: false,
		},
		{
			name: "AuthServer and AppID configured",
			config: Config{
				AuthServer: "https://example.com",
				AppID:      12345,
			},
			expected: false,
		},
		{
			name: "AuthServer and InstallationID configured",
			config: Config{
				AuthServer:     "https://example.com",
				InstallationID: 67890,
			},
			expected: false,
		},
		{
			name: "AppID and InstallationID configured",
			config: Config{
				AppID:          12345,
				InstallationID: 67890,
			},
			expected: false,
		},
		{
			name: "All fields configured",
			config: Config{
				AuthServer:     "https://example.com",
				AppID:          12345,
				InstallationID: 67890,
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.config.IsGitHubAppConfigured()
			if result != tc.expected {
				t.Errorf("Expected IsGitHubAppConfigured() to return %v, got %v", tc.expected, result)
			}
		})
	}
}

// TestLoadAndSave tests the configuration load and save functionality
// with temporary directory
func TestLoadAndSave(t *testing.T) {
	// Create a temporary directory for testing
	homeDir := t.TempDir()
	
	// Save the original environment variables
	originalHome := os.Getenv("HOME")
	originalUserProfile := os.Getenv("USERPROFILE")
	originalXdgConfig := os.Getenv("XDG_CONFIG_HOME")
	
	// Set environment variables to point to our test directory
	os.Setenv("HOME", homeDir)
	os.Setenv("USERPROFILE", homeDir)
	os.Setenv("XDG_CONFIG_HOME", homeDir)
	
	// Restore environment variables after the test
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("USERPROFILE", originalUserProfile)
		os.Setenv("XDG_CONFIG_HOME", originalXdgConfig)
	}()

	// Create a test config
	testConfig := &Config{
		AuthServer:     "https://auth.example.com",
		AppID:          12345,
		InstallationID: 67890,
	}

	// Save the config
	err := Save(testConfig)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load the config
	loadedConfig, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify the loaded config matches the saved config
	if loadedConfig.AuthServer != testConfig.AuthServer {
		t.Errorf("Expected AuthServer to be %q, got %q", testConfig.AuthServer, loadedConfig.AuthServer)
	}
	if loadedConfig.AppID != testConfig.AppID {
		t.Errorf("Expected AppID to be %d, got %d", testConfig.AppID, loadedConfig.AppID)
	}
	if loadedConfig.InstallationID != testConfig.InstallationID {
		t.Errorf("Expected InstallationID to be %d, got %d", testConfig.InstallationID, loadedConfig.InstallationID)
	}
}

// Test config path functions
func TestConfigPaths(t *testing.T) {
	// Create a temporary directory for testing
	homeDir := t.TempDir()
	
	// Save the original environment variables
	originalHome := os.Getenv("HOME")
	originalUserProfile := os.Getenv("USERPROFILE")
	originalXdgConfig := os.Getenv("XDG_CONFIG_HOME")
	
	// Set environment variables to point to our test directory
	os.Setenv("HOME", homeDir)
	os.Setenv("USERPROFILE", homeDir)
	os.Setenv("XDG_CONFIG_HOME", homeDir)
	
	// Restore environment variables after the test
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("USERPROFILE", originalUserProfile)
		os.Setenv("XDG_CONFIG_HOME", originalXdgConfig)
	}()

	// Test getConfigDir
	configDir, err := getConfigDir()
	if err != nil {
		t.Fatalf("getConfigDir returned error: %v", err)
	}
	
	// Verify the directory was created
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Errorf("Config directory was not created at %s", configDir)
	}
}

// TestLoadWithNoFile tests the Load function when no config file exists
func TestLoadWithNoFile(t *testing.T) {
	// Create a temporary directory for testing
	homeDir := t.TempDir()
	
	// Save the original environment variables
	originalHome := os.Getenv("HOME")
	originalUserProfile := os.Getenv("USERPROFILE")
	originalXdgConfig := os.Getenv("XDG_CONFIG_HOME")
	
	// Set environment variables to point to our test directory
	os.Setenv("HOME", homeDir)
	os.Setenv("USERPROFILE", homeDir)
	os.Setenv("XDG_CONFIG_HOME", homeDir)
	
	// Restore environment variables after the test
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("USERPROFILE", originalUserProfile)
		os.Setenv("XDG_CONFIG_HOME", originalXdgConfig)
	}()

	// Call Load - should create an empty config
	config, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	// Check that we got an empty config
	if config == nil {
		t.Fatal("Expected non-nil config, got nil")
	}
	if config.AuthServer != "" || config.AppID != 0 || config.InstallationID != 0 {
		t.Errorf("Expected empty config, got %+v", config)
	}
}