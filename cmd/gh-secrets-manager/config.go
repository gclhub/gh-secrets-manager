package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"gh-secrets-manager/pkg/config"

	"github.com/spf13/cobra"
)

// validConfigKeys defines the only allowed configuration keys
var validConfigKeys = map[string]bool{
	"auth-server":     true,
	"app-id":          true,
	"installation-id": true,
}

// validateConfigKey checks if a configuration key is valid
func validateConfigKey(key string) error {
	if !validConfigKeys[key] {
		return fmt.Errorf("invalid configuration key: %s. Valid keys are: auth-server, app-id, installation-id", key)
	}
	return nil
}

// validateIntegerValue validates that a value can be parsed as a positive integer
func validateIntegerValue(key, value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer", key)
	}
	if id <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return id, nil
}

// isLocalhostAddress checks if a host is a localhost address
func isLocalhostAddress(host string) bool {
	// Remove port number if present
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	// Check for localhost
	if strings.ToLower(host) == "localhost" {
		return true
	}

	// Check for loopback IP addresses
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}

	return false
}

// validateURLValue validates that a value is a valid URL with appropriate protocol
func validateURLValue(value string) error {
	parsedURL, err := url.Parse(value)
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL must include a valid host")
	}

	// Allow HTTP only for localhost addresses
	if parsedURL.Scheme == "http" {
		if !isLocalhostAddress(parsedURL.Host) {
			return fmt.Errorf("HTTP protocol is only allowed for localhost addresses")
		}
	} else if parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use HTTPS protocol (HTTP is allowed only for localhost)")
	}

	return nil
}

// validateConfigValue checks if a value is valid for the given key
func validateConfigValue(key, value string) error {
	switch key {
	case "auth-server":
		return validateURLValue(value)
	case "app-id", "installation-id":
		_, err := validateIntegerValue(key, value)
		return err
	}
	return nil
}

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage GitHub App authentication configuration",
		Long: `Manage GitHub App authentication configuration settings.
Configuration is stored in the user's config directory and is used for GitHub App authentication.`,
	}

	cmd.AddCommand(newConfigViewCmd())
	cmd.AddCommand(newConfigGetCmd())
	cmd.AddCommand(newConfigSetCmd())
	cmd.AddCommand(newConfigDeleteCmd())

	return cmd
}

func newConfigViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view",
		Short: "View current configuration",
		Long:  `Display all current configuration settings.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format config: %w", err)
			}

			fmt.Println(string(data))
			return nil
		},
	}
}

func newConfigGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Long:  `Get the value of a specific configuration key.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateConfigKey(args[0]); err != nil {
				return err
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			key := args[0]
			switch key {
			case "auth-server":
				fmt.Println(cfg.AuthServer)
			case "app-id":
				fmt.Println(cfg.AppID)
			case "installation-id":
				fmt.Println(cfg.InstallationID)
			}
			return nil
		},
	}
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long:  `Set the value of a specific configuration key.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value := args[0], args[1]
			
			if err := validateConfigKey(key); err != nil {
				return err
			}

			if err := validateConfigValue(key, value); err != nil {
				return err
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			switch key {
			case "auth-server":
				cfg.AuthServer = value
			case "app-id":
				id, _ := validateIntegerValue(key, value) // Error already checked by validateConfigValue
				cfg.AppID = id
			case "installation-id":
				id, _ := validateIntegerValue(key, value) // Error already checked by validateConfigValue
				cfg.InstallationID = id
			}

			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("Successfully set %s to %s\n", key, value)
			return nil
		},
	}
}

func newConfigDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <key>",
		Short: "Delete a configuration value",
		Long:  `Delete the value of a specific configuration key.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateConfigKey(args[0]); err != nil {
				return err
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			key := args[0]
			switch key {
			case "auth-server":
				cfg.AuthServer = ""
			case "app-id":
				cfg.AppID = 0
			case "installation-id":
				cfg.InstallationID = 0
			}

			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("Successfully deleted %s\n", key)
			return nil
		},
	}
}