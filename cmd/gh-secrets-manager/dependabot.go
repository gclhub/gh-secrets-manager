package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gh-secrets-manager/pkg/api"
	fileio "gh-secrets-manager/pkg/io"
	"github.com/google/go-github/v45/github"
	"github.com/spf13/cobra"
)

func newDependabotCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dependabot",
		Short: "Manage Dependabot secrets",
		Long:  `Manage Dependabot secrets at both organization and repository levels.`,
	}
}

func addDependabotCommands(rootCmd *cobra.Command, opts *api.ClientOptions) {
	dependabotCmd := newDependabotCmd()

	// List dependabot secrets command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List Dependabot secrets",
		Long: `List Dependabot secrets at organization or repository level.

Usage:
  # List organization Dependabot secrets
  $ gh secrets-manager dependabot list --org myorg

  # List repository Dependabot secrets
  $ gh secrets-manager dependabot list --repo owner/repo

  # List Dependabot secrets for repositories with specific property
  $ gh secrets-manager dependabot list --org myorg --property team --prop_value backend`,
		Example: `  # List all Dependabot secrets in an organization
  $ gh secrets-manager dependabot list --org myorg

  # List Dependabot secrets in a specific repository
  $ gh secrets-manager dependabot list --repo owner/repo

  # List Dependabot secrets for all frontend team repositories
  $ gh secrets-manager dependabot list --org myorg --property team --prop_value frontend`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListDependabotSecrets(cmd, opts)
		},
	}

	// Set dependabot secrets command
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Create or update Dependabot secrets",
		Long: `Create or update Dependabot secrets from command line or file input.

Input Methods:
  1. Command Line:
     Provide secret name and value directly using flags
     
  2. File Input:
     Import secrets from JSON or CSV files
     Supported formats:
     - JSON: Array of {"name": "SECRET_NAME", "value": "secret_value"}
     - CSV: Two columns with headers "name,value"

Common Use Cases:
  - NPM_TOKEN for private npm registry access
  - MAVEN_USERNAME and MAVEN_PASSWORD for private Maven repositories
  - NUGET_TOKEN for private NuGet feeds
  - DOCKER_USERNAME and DOCKER_PASSWORD for private container registries

Note: Secret values are encrypted before transmission using GitHub's public key.

Usage:
  # Set a single Dependabot secret
  $ gh secrets-manager dependabot set --org myorg --name NPM_TOKEN --value "1234567890"

  # Import Dependabot secrets from file
  $ gh secrets-manager dependabot set --org myorg --file dependabot-secrets.json

  # Set Dependabot secret for specific repositories
  $ gh secrets-manager dependabot set --org myorg --property team --prop_value backend --name DOCKER_TOKEN --value "abcdef123456"`,
		Example: `  # Set a Dependabot secret in an organization
  $ gh secrets-manager dependabot set --org myorg --name NPM_TOKEN --value "1234567890"

  # Set a Dependabot secret in a repository
  $ gh secrets-manager dependabot set --repo owner/repo --name NUGET_TOKEN --value "abcdef123456"

  # Import Dependabot secrets from JSON file
  $ gh secrets-manager dependabot set --org myorg --file dependabot-secrets.json

  # Set Dependabot secret for all backend repositories
  $ gh secrets-manager dependabot set --org myorg --property team --prop_value backend --name MAVEN_PASSWORD --value "secret123"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetDependabotSecrets(cmd, opts)
		},
	}

	// Delete dependabot secrets command
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete Dependabot secrets",
		Long: `Delete Dependabot secrets at organization or repository level.

Usage:
  # Delete an organization Dependabot secret
  $ gh secrets-manager dependabot delete --org myorg --name SECRET_NAME

  # Delete a repository Dependabot secret
  $ gh secrets-manager dependabot delete --repo owner/repo --name SECRET_NAME

  # Delete a Dependabot secret from repositories with specific property
  $ gh secrets-manager dependabot delete --org myorg --property team --prop_value backend --name NPM_TOKEN`,
		Example: `  # Delete a Dependabot secret from an organization
  $ gh secrets-manager dependabot delete --org myorg --name NPM_TOKEN

  # Delete a Dependabot secret from a repository
  $ gh secrets-manager dependabot delete --repo owner/repo --name NUGET_TOKEN

  # Delete a Dependabot secret from all frontend repositories
  $ gh secrets-manager dependabot delete --org myorg --property team --prop_value frontend --name DOCKER_PASSWORD`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeleteDependabotSecrets(cmd, opts)
		},
	}

	// Add common flags to all commands
	for _, command := range []*cobra.Command{listCmd, setCmd, deleteCmd} {
		addCommonFlags(command)
	}

	// Add specific flags for set command
	setCmd.Flags().StringP("file", "f", "", "JSON/CSV file containing secrets (format: array of {\"name\": \"SECRET_NAME\", \"value\": \"secret_value\"})")
	setCmd.Flags().String("name", "", "Secret name (e.g., NPM_TOKEN)")
	setCmd.Flags().String("value", "", "Secret value to encrypt and store")

	// Add specific flags for delete command
	deleteCmd.Flags().String("name", "", "Secret name to delete")

	// Add all commands to dependabot command
	dependabotCmd.AddCommand(listCmd, setCmd, deleteCmd)
	rootCmd.AddCommand(dependabotCmd)
}

func runListDependabotSecrets(cmd *cobra.Command, opts *api.ClientOptions) error {
	client, err := api.NewClientWithOptions(opts)
	if err != nil {
		return err
	}

	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")
	property, _ := cmd.Flags().GetString("property")
	value, _ := cmd.Flags().GetString("prop_value")

	if org != "" {
		if property != "" && value != "" {
			// List secrets for repositories matching property
			repos, err := client.ListRepositoriesByProperty(org, property, value)
			if err != nil {
				return err
			}

			var results []map[string]interface{}
			for _, repo := range repos {
				secrets, err := client.ListRepoDependabotSecrets(org, repo.GetName())
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to list Dependabot secrets for %s: %v\n", repo.GetName(), err)
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
		secrets, err := client.ListOrgDependabotSecrets(org)
		if err != nil {
			return err
		}
		return outputJSON(secrets)
	}

	if repo != "" {
		owner, repoName := splitRepo(repo)
		secrets, err := client.ListRepoDependabotSecrets(owner, repoName)
		if err != nil {
			return err
		}
		return outputJSON(secrets)
	}

	return fmt.Errorf("either --org or --repo flag must be specified")
}

func runSetDependabotSecrets(cmd *cobra.Command, opts *api.ClientOptions) error {
	client, err := api.NewClientWithOptions(opts)
	if err != nil {
		return err
	}

	file, _ := cmd.Flags().GetString("file")
	if file != "" {
		return handleDependabotFileInput(cmd, client, file)
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
				if err := client.CreateOrUpdateRepoDependabotSecret(org, repo.GetName(), secret); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to set Dependabot secret for %s: %v\n", repo.GetName(), err)
					lastErr = err
					continue
				}
			}
			return lastErr
		}

		return client.CreateOrUpdateOrgDependabotSecret(org, secret)
	}

	if repo != "" {
		owner, repoName := splitRepo(repo)
		return client.CreateOrUpdateRepoDependabotSecret(owner, repoName, secret)
	}

	return fmt.Errorf("either --org or --repo flag must be specified")
}

func runDeleteDependabotSecrets(cmd *cobra.Command, opts *api.ClientOptions) error {
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

	if org != "" {
		if property != "" && value != "" {
			// Delete secret from repositories matching property
			repos, err := client.ListRepositoriesByProperty(org, property, value)
			if err != nil {
				return err
			}

			var lastErr error
			for _, repo := range repos {
				if err := client.DeleteRepoDependabotSecret(org, repo.GetName(), name); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to delete Dependabot secret from %s: %v\n", repo.GetName(), err)
					lastErr = err
					continue
				}
			}
			return lastErr
		}

		return client.DeleteOrgDependabotSecret(org, name)
	}

	if repo != "" {
		owner, repoName := splitRepo(repo)
		return client.DeleteRepoDependabotSecret(owner, repoName, name)
	}

	return fmt.Errorf("either --org or --repo flag must be specified")
}

func handleDependabotFileInput(cmd *cobra.Command, client *api.Client, filePath string) error {
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
					if err := client.CreateOrUpdateRepoDependabotSecret(org, repo.GetName(), encSecret); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: Failed to set Dependabot secret %s for %s: %v\n", secret.Name, repo.GetName(), err)
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
			if err := client.CreateOrUpdateOrgDependabotSecret(org, encSecret); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to set organization Dependabot secret %s: %v\n", secret.Name, err)
				lastErr = err
				continue
			}
		}
		return lastErr
	}

	if repo != "" {
		owner, repoName := splitRepo(repo)
		for _, secret := range secrets {
			encSecret := &github.EncryptedSecret{
				Name:           secret.Name,
				EncryptedValue: secret.Value,
			}
			if err := client.CreateOrUpdateRepoDependabotSecret(owner, repoName, encSecret); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to set repository Dependabot secret %s: %v\n", secret.Name, err)
				lastErr = err
				continue
			}
		}
		return lastErr
	}

	return fmt.Errorf("either --org or --repo flag must be specified")
}
