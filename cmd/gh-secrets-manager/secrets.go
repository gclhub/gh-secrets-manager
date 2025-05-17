package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gh-secrets-manager/pkg/api"
	fileio "gh-secrets-manager/pkg/io"
	"github.com/google/go-github/v45/github"
	"github.com/spf13/cobra"
)

func newSecretsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "secrets",
		Short: "Manage GitHub Actions secrets",
		Long:  `Manage GitHub Actions secrets at both organization and repository levels.`,
	}
}

func addSecretCommands(rootCmd *cobra.Command, opts *api.ClientOptions) {
	secretsCmd := newSecretsCmd()

	// List secrets command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List GitHub Actions secrets",
		Long: `List GitHub Actions secrets at organization, repository, or environment level.

Usage:
  # List organization secrets
  $ gh secrets-manager secrets list --org myorg

  # List repository secrets
  $ gh secrets-manager secrets list --repo owner/repo

  # List secrets for repositories with specific property
  $ gh secrets-manager secrets list --org myorg --property team --prop_value backend

  # List environment secrets
  $ gh secrets-manager secrets list --repo owner/repo --environment prod`,
		Example: `  # List all secrets in an organization
  $ gh secrets-manager secrets list --org myorg

  # List secrets in a specific repository
  $ gh secrets-manager secrets list --repo owner/repo

  # List secrets for all frontend team repositories
  $ gh secrets-manager secrets list --org myorg --property team --prop_value frontend

  # List secrets in an environment
  $ gh secrets-manager secrets list --repo owner/repo --environment prod`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListSecrets(cmd, opts)
		},
	}

	// Set secrets command
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Create or update GitHub Actions secrets",
		Long: `Create or update GitHub Actions secrets from command line or file input.

Input Methods:
  1. Command Line:
     Provide secret name and value directly using flags
     
  2. File Input:
     Import secrets from JSON or CSV files
     Supported formats:
     - JSON: Array of {"name": "SECRET_NAME", "value": "secret_value"}
     - CSV: Two columns with headers "name,value"

Security Notes:
  - All secrets are encrypted using libsodium sealed boxes
  - Values are encrypted before being sent to GitHub
  - Secrets are only decrypted during workflow execution
  - Organization secrets can be restricted to specific repositories

Usage:
  # Set a single secret
  $ gh secrets-manager secrets set --org myorg --name API_KEY --value "1234567890"

  # Import secrets from file
  $ gh secrets-manager secrets set --org myorg --file secrets.json

  # Set secret for specific repositories
  $ gh secrets-manager secrets set --org myorg --property team --prop_value backend --name DB_PASSWORD --value "secretpass"

  # Set secret in an environment
  $ gh secrets-manager secrets set --repo owner/repo --environment prod --name API_KEY --value "1234567890"`,
		Example: `  # Set a secret in an organization
  $ gh secrets-manager secrets set --org myorg --name AWS_KEY --value "AKIAXXXXXXXX"

  # Set a secret in a repository
  $ gh secrets-manager secrets set --repo owner/repo --name DEPLOY_TOKEN --value "ghp_XXXXXX"

  # Import secrets from JSON file
  $ gh secrets-manager secrets set --org myorg --file secrets.json

  # Set secret for all backend repositories
  $ gh secrets-manager secrets set --org myorg --property team --prop_value backend --name DB_PASSWORD --value "secretpass"

  # Set secret in an environment
  $ gh secrets-manager secrets set --repo owner/repo --environment prod --name API_KEY --value "1234567890"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetSecrets(cmd, opts)
		},
	}

	// Delete secrets command
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete GitHub Actions secrets",
		Long: `Delete GitHub Actions secrets at organization, repository, or environment level.

Usage:
  # Delete an organization secret
  $ gh secrets-manager secrets delete --org myorg --name SECRET_NAME

  # Delete a repository secret
  $ gh secrets-manager secrets delete --repo owner/repo --name SECRET_NAME

  # Delete a secret from repositories with specific property
  $ gh secrets-manager secrets delete --org myorg --property team --prop_value backend --name DB_PASSWORD

  # Delete a secret from an environment
  $ gh secrets-manager secrets delete --repo owner/repo --environment prod --name SECRET_NAME`,
		Example: `  # Delete a secret from an organization
  $ gh secrets-manager secrets delete --org myorg --name AWS_KEY

  # Delete a secret from a repository
  $ gh secrets-manager secrets delete --repo owner/repo --name DEPLOY_TOKEN

  # Delete a secret from all frontend repositories
  $ gh secrets-manager secrets delete --org myorg --property team --prop_value frontend --name API_KEY

  # Delete a secret from an environment
  $ gh secrets-manager secrets delete --repo owner/repo --environment prod --name SECRET_NAME`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeleteSecrets(cmd, opts)
		},
	}

	// Add common flags to all commands
	for _, command := range []*cobra.Command{listCmd, setCmd, deleteCmd} {
		addCommonFlags(command)
	}

	// Add specific flags for set command
	setCmd.Flags().StringP("file", "f", "", "JSON/CSV file containing secrets (format: array of {\"name\": \"SECRET_NAME\", \"value\": \"secret_value\"})")
	setCmd.Flags().String("name", "", "Secret name (e.g., API_KEY)")
	setCmd.Flags().String("value", "", "Secret value to encrypt and store")
	setCmd.Flags().String("environment", "", "GitHub Actions environment name")

	// Add specific flags for delete command
	deleteCmd.Flags().String("name", "", "Secret name to delete")
	deleteCmd.Flags().String("environment", "", "GitHub Actions environment name")

	// Add environment flag to list command
	listCmd.Flags().String("environment", "", "GitHub Actions environment name")

	// Add all commands to secrets command
	secretsCmd.AddCommand(listCmd, setCmd, deleteCmd)
	rootCmd.AddCommand(secretsCmd)
}

func addCommonFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("org", "o", "", "GitHub organization name")
	cmd.Flags().StringP("repo", "r", "", "GitHub repository name")
	cmd.Flags().String("property", "", "Custom property name for filtering repositories")
	cmd.Flags().String("prop_value", "", "Custom property value for filtering repositories")
}

func runListSecrets(cmd *cobra.Command, opts *api.ClientOptions) error {
	client, err := api.NewClientWithOptions(opts)
	if err != nil {
		return err
	}

	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")
	property, _ := cmd.Flags().GetString("property")
	value, _ := cmd.Flags().GetString("prop_value")
	environment, _ := cmd.Flags().GetString("environment")

	if org != "" {
		if property != "" && value != "" {
			// List secrets for repositories matching property
			repos, err := client.ListRepositoriesByProperty(org, property, value)
			if err != nil {
				return err
			}

			var results []map[string]interface{}
			for _, repo := range repos {
				secrets, err := client.ListRepoSecrets(org, repo.GetName())
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to list secrets for %s: %v\n", repo.GetName(), err)
					continue
				}
				results = append(results, map[string]interface{}{
					"repository": repo.GetName(),
					"secrets":    secrets,
				})
			}
			return outputJSON(results)
		}

		// List organization secrets
		secrets, err := client.ListOrgSecrets(org)
		if err != nil {
			return err
		}
		return outputJSON(secrets)
	}

	if repo != "" {
		owner, repoName := splitRepo(repo)
		if environment != "" {
			secrets, err := client.ListEnvironmentSecrets(owner, repoName, environment)
			if err != nil {
				return err
			}
			return outputJSON(secrets)
		}

		secrets, err := client.ListRepoSecrets(owner, repoName)
		if err != nil {
			return err
		}
		return outputJSON(secrets)
	}

	return fmt.Errorf("either --org or --repo flag must be specified")
}

func runSetSecrets(cmd *cobra.Command, opts *api.ClientOptions) error {
	client, err := api.NewClientWithOptions(opts)
	if err != nil {
		return err
	}

	file, _ := cmd.Flags().GetString("file")
	if file != "" {
		return handleFileInput(cmd, client, file)
	}

	name, _ := cmd.Flags().GetString("name")
	value, _ := cmd.Flags().GetString("value")
	if name == "" || value == "" {
		return fmt.Errorf("both --name and --value flags are required when not using a file")
	}

	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")
	property, _ := cmd.Flags().GetString("property")
	propValue, _ := cmd.Flags().GetString("prop_value")
	environment, _ := cmd.Flags().GetString("environment")

	secret := &github.EncryptedSecret{
		Name:           name,
		EncryptedValue: value,
	}

	if org != "" {
		if property != "" && propValue != "" {
			// Set secret for repositories matching property
			repos, err := client.ListRepositoriesByProperty(org, property, propValue)
			if err != nil {
				return err
			}

			var lastErr error
			for _, repo := range repos {
				if err := client.CreateOrUpdateRepoSecret(org, repo.GetName(), secret); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to set secret for %s: %v\n", repo.GetName(), err)
					lastErr = err
					continue
				}
			}
			return lastErr
		}

		return client.CreateOrUpdateOrgSecret(org, secret)
	}

	if repo != "" {
		owner, repoName := splitRepo(repo)
		if environment != "" {
			return client.CreateOrUpdateEnvironmentSecret(owner, repoName, environment, secret)
		}
		return client.CreateOrUpdateRepoSecret(owner, repoName, secret)
	}

	return fmt.Errorf("either --org or --repo flag must be specified")
}

func runDeleteSecrets(cmd *cobra.Command, opts *api.ClientOptions) error {
	client, err := api.NewClientWithOptions(opts)
	if err != nil {
		return err
	}

	name, _ := cmd.Flags().GetString("name")
	if name == "" {
		return fmt.Errorf("--name flag is required")
	}

	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")
	property, _ := cmd.Flags().GetString("property")
	value, _ := cmd.Flags().GetString("prop_value")
	environment, _ := cmd.Flags().GetString("environment")

	if org != "" {
		if property != "" && value != "" {
			// Delete secret from repositories matching property
			repos, err := client.ListRepositoriesByProperty(org, property, value)
			if err != nil {
				return err
			}

			var lastErr error
			for _, repo := range repos {
				if err := client.DeleteRepoSecret(org, repo.GetName(), name); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to delete secret from %s: %v\n", repo.GetName(), err)
					lastErr = err
					continue
				}
			}
			return lastErr
		}

		return client.DeleteOrgSecret(org, name)
	}

	if repo != "" {
		owner, repoName := splitRepo(repo)
		if environment != "" {
			return client.DeleteEnvironmentSecret(owner, repoName, environment, name)
		}
		return client.DeleteRepoSecret(owner, repoName, name)
	}

	return fmt.Errorf("either --org or --repo flag must be specified")
}

func handleFileInput(cmd *cobra.Command, client *api.Client, filePath string) error {
	ext := filepath.Ext(filePath)
	var secrets []fileio.SecretData
	var err error

	switch ext {
	case ".json":
		secrets, err = fileio.ReadJSONSecrets(filePath)
	case ".csv":
		secrets, err = fileio.ReadCSVSecrets(filePath)
	default:
		return fmt.Errorf("unsupported file format: %s", ext)
	}

	if err != nil {
		return err
	}

	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")
	property, _ := cmd.Flags().GetString("property")
	value, _ := cmd.Flags().GetString("prop_value")
	environment, _ := cmd.Flags().GetString("environment")

	var lastErr error
	if org != "" {
		if property != "" && value != "" {
			// Set secrets for repositories matching property
			repos, err := client.ListRepositoriesByProperty(org, property, value)
			if err != nil {
				return err
			}

			for _, repo := range repos {
				for _, secret := range secrets {
					encSecret := &github.EncryptedSecret{
						Name:           secret.Name,
						EncryptedValue: secret.Value,
					}
					if err := client.CreateOrUpdateRepoSecret(org, repo.GetName(), encSecret); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: Failed to set secret %s for %s: %v\n", secret.Name, repo.GetName(), err)
						lastErr = err
						continue
					}
				}
			}
			return lastErr
		}

		// Set organization secrets
		for _, secret := range secrets {
			encSecret := &github.EncryptedSecret{
				Name:           secret.Name,
				EncryptedValue: secret.Value,
			}
			if err := client.CreateOrUpdateOrgSecret(org, encSecret); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to set organization secret %s: %v\n", secret.Name, err)
				lastErr = err
				continue
			}
		}
		return lastErr
	}

	if repo != "" {
		owner, repoName := splitRepo(repo)
		if environment != "" {
			for _, secret := range secrets {
				encSecret := &github.EncryptedSecret{
					Name:           secret.Name,
					EncryptedValue: secret.Value,
				}
				if err := client.CreateOrUpdateEnvironmentSecret(owner, repoName, environment, encSecret); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to set environment secret %s: %v\n", secret.Name, err)
					lastErr = err
					continue
				}
			}
			return lastErr
		}

		for _, secret := range secrets {
			encSecret := &github.EncryptedSecret{
				Name:           secret.Name,
				EncryptedValue: secret.Value,
			}
			if err := client.CreateOrUpdateRepoSecret(owner, repoName, encSecret); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to set repository secret %s: %v\n", secret.Name, err)
				lastErr = err
				continue
			}
		}
		return lastErr
	}

	return fmt.Errorf("either --org or --repo flag must be specified")
}

func splitRepo(repo string) (string, string) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return "", repo
	}
	return parts[0], parts[1]
}

func outputJSON(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}