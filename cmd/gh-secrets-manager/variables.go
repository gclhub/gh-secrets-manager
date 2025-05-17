package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gh-secrets-manager/pkg/api"
	fileio "gh-secrets-manager/pkg/io"
	"github.com/spf13/cobra"
)

func newVariablesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "variables",
		Short: "Manage GitHub Actions variables",
		Long:  `Manage GitHub Actions variables at both organization and repository levels.`,
	}
}

func addVariableCommands(rootCmd *cobra.Command, opts *api.ClientOptions) {
	variablesCmd := newVariablesCmd()

	// List variables command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List GitHub Actions variables",
		Long: `List GitHub Actions variables at organization or repository level.

Usage:
  # List organization variables
  $ gh secrets-manager variables list --org myorg

  # List repository variables
  $ gh secrets-manager variables list --repo owner/repo

  # List environment variables
  $ gh secrets-manager variables list --repo owner/repo --environment prod

  # List variables for repositories with specific property
  $ gh secrets-manager variables list --org myorg --property team --prop_value backend`,
		Example: `  # List all variables in an organization
  $ gh secrets-manager variables list --org myorg

  # List variables in a repository
  $ gh secrets-manager variables list --repo owner/repo

  # List variables in an environment
  $ gh secrets-manager variables list --repo owner/repo --environment prod

  # List variables in all frontend repos
  $ gh secrets-manager variables list --org myorg --property team --prop_value frontend`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListVariables(cmd, opts)
		},
	}

	// Set variables command
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Create or update GitHub Actions variables",
		Long: `Create or update GitHub Actions variables from command line or file input.

Input Methods:
  1. Command Line:
     Provide variable name and value directly using flags
     
  2. File Input:
     Import variables from JSON or CSV files
     Supported formats:
     - JSON: Array of {"name": "VAR_NAME", "value": "var_value"}
     - CSV: Two columns with headers "name,value"

Usage:
  # Set a single variable
  $ gh secrets-manager variables set --org myorg --name CONFIG_ENV --value "production"

  # Import variables from file
  $ gh secrets-manager variables set --org myorg --file variables.json

  # Set variable for specific repositories
  $ gh secrets-manager variables set --org myorg --property team --prop_value backend --name DB_HOST --value "db.example.com"

  # Set environment variable
  $ gh secrets-manager variables set --repo owner/repo --environment prod --name API_URL --value "api.example.com"`,
		Example: `  # Set a variable in an organization
  $ gh secrets-manager variables set --org myorg --name NODE_ENV --value "production"

  # Set a variable in a repository
  $ gh secrets-manager variables set --repo owner/repo --name PORT --value "8080"

  # Import variables from JSON file
  $ gh secrets-manager variables set --org myorg --file variables.json

  # Set variable for all backend repositories
  $ gh secrets-manager variables set --org myorg --property team --prop_value backend --name LOG_LEVEL --value "info"

  # Set variable in an environment
  $ gh secrets-manager variables set --repo owner/repo --environment prod --name API_URL --value "api.example.com"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetVariables(cmd, opts)
		},
	}

	// Delete variables command
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete GitHub Actions variables",
		Long: `Delete GitHub Actions variables at organization, repository, or environment level.

Usage:
  # Delete an organization variable
  $ gh secrets-manager variables delete --org myorg --name VAR_NAME

  # Delete a repository variable
  $ gh secrets-manager variables delete --repo owner/repo --name VAR_NAME

  # Delete a variable from repositories with specific property
  $ gh secrets-manager variables delete --org myorg --property team --prop_value backend --name DB_HOST

  # Delete an environment variable
  $ gh secrets-manager variables delete --repo owner/repo --environment prod --name VAR_NAME`,
		Example: `  # Delete a variable from an organization
  $ gh secrets-manager variables delete --org myorg --name NODE_ENV

  # Delete a variable from a repository
  $ gh secrets-manager variables delete --repo owner/repo --name PORT

  # Delete a variable from all frontend repositories
  $ gh secrets-manager variables delete --org myorg --property team --prop_value frontend --name API_URL

  # Delete a variable from an environment
  $ gh secrets-manager variables delete --repo owner/repo --environment prod --name API_URL`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeleteVariables(cmd, opts)
		},
	}

	// Add common flags to all commands
	for _, command := range []*cobra.Command{listCmd, setCmd, deleteCmd} {
		addCommonFlags(command)
		command.Flags().String("environment", "", "GitHub Actions environment name")
	}

	// Add specific flags for set command
	setCmd.Flags().StringP("file", "f", "", "JSON/CSV file containing variables")
	setCmd.Flags().String("name", "", "Variable name")
	setCmd.Flags().String("value", "", "Variable value")

	// Add specific flag for delete command
	deleteCmd.Flags().String("name", "", "Variable name to delete")

	// Add all commands to variables command
	variablesCmd.AddCommand(listCmd, setCmd, deleteCmd)
	rootCmd.AddCommand(variablesCmd)
}

func runListVariables(cmd *cobra.Command, opts *api.ClientOptions) error {
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
			// List variables for repositories matching property
			repos, err := client.ListRepositoriesByProperty(org, property, value)
			if err != nil {
				return err
			}

			var results []map[string]interface{}
			for _, repo := range repos {
				variables, err := client.ListRepoVariables(org, repo.GetName())
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to list variables for %s: %v\n", repo.GetName(), err)
					continue
				}
				results = append(results, map[string]interface{}{
					"repository": repo.GetName(),
					"variables": variables,
				})
			}
			return outputJSON(results)
		}

		variables, err := client.ListOrgVariables(org)
		if err != nil {
			return err
		}
		return outputJSON(variables)
	}

	if repo != "" {
		owner, repoName := splitRepo(repo)
		if environment != "" {
			// List environment variables
			variables, err := client.ListEnvironmentVariables(owner, repoName, environment)
			if err != nil {
				return err
			}
			return outputJSON(variables)
		}

		variables, err := client.ListRepoVariables(owner, repoName)
		if err != nil {
			return err
		}
		return outputJSON(variables)
	}

	return fmt.Errorf("either --org or --repo flag must be specified")
}

func runSetVariables(cmd *cobra.Command, opts *api.ClientOptions) error {
	client, err := api.NewClientWithOptions(opts)
	if err != nil {
		return err
	}

	file, _ := cmd.Flags().GetString("file")
	if file != "" {
		return handleVariableFileInput(cmd, client, file)
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

	variable := &api.Variable{
		Name:  name,
		Value: value,
	}

	if org != "" {
		if property != "" && propValue != "" {
			// Set variable for repositories matching property
			repos, err := client.ListRepositoriesByProperty(org, property, propValue)
			if err != nil {
				return err
			}

			var lastErr error
			for _, repo := range repos {
				if err := client.CreateOrUpdateRepoVariable(org, repo.GetName(), variable); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to set variable for %s: %v\n", repo.GetName(), err)
					lastErr = err
					continue
				}
			}
			return lastErr
		}

		return client.CreateOrUpdateOrgVariable(org, variable)
	}

	if repo != "" {
		owner, repoName := splitRepo(repo)
		if environment != "" {
			// Set environment variable
			return client.CreateOrUpdateEnvironmentVariable(owner, repoName, environment, variable)
		}
		return client.CreateOrUpdateRepoVariable(owner, repoName, variable)
	}

	return fmt.Errorf("either --org or --repo flag must be specified")
}

func runDeleteVariables(cmd *cobra.Command, opts *api.ClientOptions) error {
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
			// Delete variable from repositories matching property
			repos, err := client.ListRepositoriesByProperty(org, property, value)
			if err != nil {
				return err
			}

			var lastErr error
			for _, repo := range repos {
				if err := client.DeleteRepoVariable(org, repo.GetName(), name); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to delete variable from %s: %v\n", repo.GetName(), err)
					lastErr = err
					continue
				}
			}
			return lastErr
		}

		return client.DeleteOrgVariable(org, name)
	}

	if repo != "" {
		owner, repoName := splitRepo(repo)
		if environment != "" {
			// Delete environment variable
			return client.DeleteEnvironmentVariable(owner, repoName, environment, name)
		}
		return client.DeleteRepoVariable(owner, repoName, name)
	}

	return fmt.Errorf("either --org or --repo flag must be specified")
}

func handleVariableFileInput(cmd *cobra.Command, client *api.Client, filePath string) error {
	ext := filepath.Ext(filePath)
	var variables []fileio.SecretData // Reuse SecretData type since structure is same
	var err error

	switch ext {
	case ".json":
		variables, err = fileio.ReadJSONSecrets(filePath)
	case ".csv":
		variables, err = fileio.ReadCSVSecrets(filePath)
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
			// Set variables for repositories matching property
			repos, err := client.ListRepositoriesByProperty(org, property, value)
			if err != nil {
				return err
			}

			for _, repo := range repos {
				for _, v := range variables {
					variable := &api.Variable{
						Name:  v.Name,
						Value: v.Value,
					}
					if err := client.CreateOrUpdateRepoVariable(org, repo.GetName(), variable); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: Failed to set variable %s for %s: %v\n", v.Name, repo.GetName(), err)
						lastErr = err
						continue
					}
				}
			}
			return lastErr
		}

		// Set organization variables
		for _, v := range variables {
			variable := &api.Variable{
				Name:  v.Name,
				Value: v.Value,
			}
			if err := client.CreateOrUpdateOrgVariable(org, variable); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to set organization variable %s: %v\n", v.Name, err)
				lastErr = err
				continue
			}
		}
		return lastErr
	}

	if repo != "" {
		owner, repoName := splitRepo(repo)
		if environment != "" {
			// Set environment variables
			for _, v := range variables {
				variable := &api.Variable{
					Name:  v.Name,
					Value: v.Value,
				}
				if err := client.CreateOrUpdateEnvironmentVariable(owner, repoName, environment, variable); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to set environment variable %s: %v\n", v.Name, err)
					lastErr = err
					continue
				}
			}
			return lastErr
		}

		// Set repository variables
		for _, v := range variables {
			variable := &api.Variable{
				Name:  v.Name,
				Value: v.Value,
			}
			if err := client.CreateOrUpdateRepoVariable(owner, repoName, variable); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to set repository variable %s: %v\n", v.Name, err)
				lastErr = err
				continue
			}
		}
		return lastErr
	}

	return fmt.Errorf("either --org or --repo flag must be specified")
}