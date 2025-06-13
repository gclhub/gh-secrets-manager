# gh-secrets-manager

A GitHub CLI extension for managing GitHub Actions secrets and variables, and Dependabot secrets at both organization and repository levels.

## Features

- Manage GitHub Actions secrets at organization and repository levels
- Handle Dependabot secrets
- Manage GitHub Actions variables
- Support for both public and private repositories
- Secure secret value handling
- Batch operations support

## Quick Start

1. **Install the extension**:
```bash
gh extension install gclhub/gh-secrets-manager
```

2. **Choose an authentication method**:
   - [GitHub App Authentication](docs/AUTH_SERVER.md) (Recommended for organizations)
   - Personal Access Token: Just run `gh auth login`

3. **Try some commands**:
```bash
# List organization secrets
gh secrets-manager secrets list --org myorg

# Set a repository secret
gh secrets-manager secrets set --repo owner/repo --name API_KEY --value "mysecret"

# Set an environment variable
gh secrets-manager variables set --repo owner/repo --name ENV --value "production"
```

See [Authentication](#authentication) and [Configuration](#configuration) for detailed setup.

## Authentication

This extension supports two authentication methods:

1. **GitHub App Authentication** (Recommended):
   - Enhanced security with temporary access tokens
   - Team-based access controls
   - Automatic token rotation
   - See [Auth Server Documentation](docs/AUTH_SERVER.md) for setup

2. **Personal Access Token** (Fallback):
   - Uses your GitHub CLI authentication
   - No additional configuration needed
   - Less secure than GitHub App authentication

The tool automatically uses GitHub App authentication when configured, falling back to PAT only if:
- GitHub App configuration is missing or incomplete
- There's an error loading the configuration

## Configuration

### Managing Configuration Settings

The `config` command allows you to view and modify settings:

```bash
# View all current settings
gh secrets-manager config view

# View specific setting
gh secrets-manager config get auth-server

# Set a configuration value
gh secrets-manager config set auth-server https://your-auth-server.example.com

# Delete a configuration value
gh secrets-manager config delete auth-server
```

### Configuration Storage

Configuration is stored in:
- macOS: `~/Library/Application Support/gh/secrets-manager/config.json`
- Linux: `~/.config/gh/secrets-manager/config.json`
- Windows: `%APPDATA%\gh\secrets-manager\config.json`

File permissions are set to 0644 to ensure secure access.

For detailed information about setting up and running the authentication server, see [Auth Server Documentation](docs/AUTH_SERVER.md).

## Troubleshooting

- Ensure you have the latest version of the GitHub CLI installed
- Check that you have the necessary permissions in the organization/repository
- For detailed logs, add the `--verbose` flag to any command

## Support

- [Report Issues](https://github.com/gclhub/gh-secrets-manager/issues)
- [Contributing Guide](docs/CONTRIBUTING.md)
- [Security Policy](docs/SECURITY.md)

## License

MIT License - see [LICENSE](LICENSE) for details.

## Advanced Usage

### Managing Secrets

```bash
# List organization secrets
gh secrets-manager secrets list --org myorg

# List repository secrets
gh secrets-manager secrets list --repo owner/repo

# List secrets for repositories with specific property
gh secrets-manager secrets list --org myorg --property team --prop_value backend

# Create/update organization secret
gh secrets-manager secrets set --org myorg --name API_KEY --value "secret123"

# Create/update repository secret
gh secrets-manager secrets set --repo owner/repo --name DB_PASS --value "dbpass"

# Import secrets from file
gh secrets-manager secrets set --org myorg --file secrets.json

# Set secret for specific repositories
gh secrets-manager secrets set --org myorg --property team --prop_value frontend --name NPM_TOKEN --value "npmtoken"

# Delete secret
gh secrets-manager secrets delete --org myorg --name API_KEY
gh secrets-manager secrets delete --repo owner/repo --name DB_PASS
```

### Managing Environment Secrets

You can manage secrets specific to GitHub Actions environments within a repository:

```bash
# List environment secrets
gh secrets-manager secrets list --repo owner/repo --environment prod

# Create/update environment secret
gh secrets-manager secrets set --repo owner/repo --environment prod --name SECRET_NAME --value "secret123"

# Delete environment secret
gh secrets-manager secrets delete --repo owner/repo --environment prod --name SECRET_NAME

# Import multiple environment secrets from file
gh secrets-manager secrets set --repo owner/repo --environment prod --file env-secrets.json
```

### Managing Variables

```bash
# List organization variables
gh secrets-manager variables list --org myorg

# List repository variables
gh secrets-manager variables list --repo owner/repo

# List variables for repositories with specific property
gh secrets-manager variables list --org myorg --property team --prop_value backend

# Create/update organization variable
gh secrets-manager variables set --org myorg --name VAR_NAME --value "value"

# Create/update repository variable
gh secrets-manager variables set --repo owner/repo --name VAR_NAME --value "value"

# Set variable for specific repositories
gh secrets-manager variables set --org myorg --property team --prop_value frontend --name CONFIG_ENV --value "production"

# Import variables from file
gh secrets-manager variables set --org myorg --file variables.json
gh secrets-manager variables set --repo owner/repo --file variables.csv

# Delete variable
gh secrets-manager variables delete --org myorg --name VAR_NAME
gh secrets-manager variables delete --repo owner/repo --name VAR_NAME
```

### Managing Environment Variables

Similarly, you can manage environment-specific variables:

```bash
# List environment variables
gh secrets-manager variables list --repo owner/repo --environment prod

# Create/update environment variable
gh secrets-manager variables set --repo owner/repo --environment prod --name VAR_NAME --value "value123"

# Delete environment variable
gh secrets-manager variables delete --repo owner/repo --environment prod --name VAR_NAME

# Import multiple environment variables from file
gh secrets-manager variables set --repo owner/repo --environment prod --file env-variables.json
```

### Managing Dependabot Secrets

```bash
# List organization Dependabot secrets
gh secrets-manager dependabot list --org myorg

# List repository Dependabot secrets
gh secrets-manager dependabot list --repo owner/repo

# List Dependabot secrets for repositories with specific property
gh secrets-manager dependabot list --org myorg --property team --prop_value backend

# Create/update organization Dependabot secret
gh secrets-manager dependabot set --org myorg --name NPM_TOKEN --value "npmtoken"

# Create/update repository Dependabot secret
gh secrets-manager dependabot set --repo owner/repo --name DOCKER_TOKEN --value "dockertoken"

# Set Dependabot secret for specific repositories
gh secrets-manager dependabot set --org myorg --property team --prop_value backend --name MAVEN_PASSWORD --value "mavenpass"

# Import Dependabot secrets from file
gh secrets-manager dependabot set --org myorg --file dependabot-secrets.json
gh secrets-manager dependabot set --repo owner/repo --file dependabot-secrets.csv

# Delete Dependabot secret
gh secrets-manager dependabot delete --org myorg --name NPM_TOKEN
gh secrets-manager dependabot delete --repo owner/repo --name DOCKER_TOKEN
```

## Input File Formats

### JSON

```json
{
  "SECRET_NAME": "secret_value",
  "ANOTHER_SECRET": "another_value"
}
```

or

```json
[
  {
    "name": "SECRET_NAME",
    "value": "secret_value"
  },
  {
    "name": "ANOTHER_SECRET",
    "value": "another_value"
  }
]
```

### CSV

```csv
name,value
SECRET_NAME,secret_value
ANOTHER_SECRET,another_value
```

## Repository Property Filtering

The `--property` and `--prop_value` flags allow you to target multiple repositories based on GitHub custom repository properties.

First, define custom properties in your organization's settings:
1. Go to your organization's settings
2. Navigate to Repository defaults > Properties
3. Add custom properties like `team`, `environment`, `service-tier`, etc.
4. Assign property values to your repositories

Then use these properties to target specific repositories:

```bash
# Set secrets for all backend team repositories
gh secrets-manager secrets set --org myorg --property team --prop_value backend --name DB_PASSWORD --value "secret123"

# Update config for all production tier repositories
gh secrets-manager variables set --org myorg --property service-tier --prop_value production --name API_URL --value "prod.api.example.com"

# Set variables for staging repositories
gh secrets-manager variables set --org myorg --property environment --prop_value staging --name LOG_LEVEL --value "debug"
```

Note: This feature requires that you have defined custom properties in your organization's settings and assigned values to repositories.